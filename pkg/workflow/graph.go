package workflow

import (
	"fmt"
	"strings"

	"github.com/solo-ai/solo/pkg/types"
)

// Graph manages the workflow DAG and topological execution order.
type Graph struct {
	nodes    map[string]*types.WorkflowNode
	adjList  map[string][]string // node ID -> dependent node IDs
	entryID  string
}

// NewGraph creates a workflow graph from workflow definition.
func NewGraph(wf *types.Workflow) (*Graph, error) {
	g := &Graph{
		nodes:   make(map[string]*types.WorkflowNode),
		adjList: make(map[string][]string),
		entryID: wf.EntryPoint,
	}

	for i := range wf.Nodes {
		node := &wf.Nodes[i]
		g.nodes[node.ID] = node
		for _, in := range node.Inputs {
			g.adjList[in] = append(g.adjList[in], node.ID)
		}
	}

	if _, ok := g.nodes[g.entryID]; !ok {
		return nil, fmt.Errorf("entry point node %q not found", g.entryID)
	}

	return g, nil
}

// TopologicalOrder returns node IDs in execution order (dependencies first).
func (g *Graph) TopologicalOrder() ([]string, error) {
	inDegree := make(map[string]int)
	for id := range g.nodes {
		inDegree[id] = 0
	}
	for _, deps := range g.adjList {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		for _, next := range g.adjList[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(g.nodes) {
		return nil, fmt.Errorf("workflow contains cycle")
	}

	return order, nil
}

// Dependencies returns the input node IDs for a given node.
func (g *Graph) Dependencies(nodeID string) ([]string, error) {
	n, ok := g.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %q not found", nodeID)
	}
	return n.Inputs, nil
}

// Node returns the workflow node by ID.
func (g *Graph) Node(id string) (*types.WorkflowNode, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

// EntryPoint returns the entry node ID.
func (g *Graph) EntryPoint() string {
	return g.entryID
}

// AllNodes returns all node IDs.
func (g *Graph) AllNodes() []string {
	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	return ids
}

// String returns a simple textual representation of the graph.
func (g *Graph) String() string {
	var sb strings.Builder
	sb.WriteString("Workflow Graph:\n")
	for id, n := range g.nodes {
		sb.WriteString(fmt.Sprintf("  %s (%s) <- %v\n", id, n.Name, n.Inputs))
	}
	return sb.String()
}
