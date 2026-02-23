package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/solo-ai/solo/pkg/types"
	"github.com/solo-ai/solo/pkg/workflow"
)

// Orchestrator coordinates reasoning pipeline execution.
// Replaces Ray's job orchestration with a simpler, Go-native approach.
type Orchestrator struct {
	graph      *workflow.Graph
	executors  map[string]types.TaskExecutor
	results    map[string]*types.Task
	resultsMu  sync.RWMutex
	maxWorkers int
}

// Config holds orchestrator configuration.
type Config struct {
	MaxWorkers int
}

// New creates an orchestrator for the given workflow.
func New(wf *types.Workflow, executors map[string]types.TaskExecutor, cfg *Config) (*Orchestrator, error) {
	g, err := workflow.NewGraph(wf)
	if err != nil {
		return nil, err
	}

	maxWorkers := 4
	if cfg != nil && cfg.MaxWorkers > 0 {
		maxWorkers = cfg.MaxWorkers
	}

	return &Orchestrator{
		graph:      g,
		executors:  executors,
		results:    make(map[string]*types.Task),
		maxWorkers: maxWorkers,
	}, nil
}

// Run executes the workflow with the given initial payload.
func (o *Orchestrator) Run(ctx context.Context, initialPayload map[string]interface{}) (map[string]interface{}, error) {
	order, err := o.graph.TopologicalOrder()
	if err != nil {
		return nil, fmt.Errorf("topological order: %w", err)
	}

	for _, nodeID := range order {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		node, ok := o.graph.Node(nodeID)
		if !ok {
			return nil, fmt.Errorf("node %q not found", nodeID)
		}

		executor, ok := o.executors[node.Executor]
		if !ok {
			return nil, fmt.Errorf("executor %q not found for node %q", node.Executor, nodeID)
		}

		// Build payload from dependencies
		payload := make(map[string]interface{})
		if len(node.Inputs) == 0 {
			payload = initialPayload
		} else {
			for _, inID := range node.Inputs {
				o.resultsMu.RLock()
				r, ok := o.results[inID]
				o.resultsMu.RUnlock()
				if !ok {
					return nil, fmt.Errorf("dependency %q not yet executed", inID)
				}
				payload[inID] = r.Result
			}
		}

		task := &types.Task{
			ID:          uuid.New().String(),
			Name:        node.Name,
			Status:      types.TaskStatusPending,
			Payload:     payload,
			Dependencies: node.Inputs,
		}

		result, err := executor.Execute(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("node %q: %w", nodeID, err)
		}

		o.resultsMu.Lock()
		o.results[nodeID] = result
		o.resultsMu.Unlock()
	}

	// Return final output (from last node in topological order)
	lastID := order[len(order)-1]
	o.resultsMu.RLock()
	r := o.results[lastID]
	o.resultsMu.RUnlock()

	if r == nil {
		return nil, fmt.Errorf("no result from final node %q", lastID)
	}
	return r.Result, nil
}
