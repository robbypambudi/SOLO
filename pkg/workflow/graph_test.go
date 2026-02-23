package workflow

import (
	"testing"

	"github.com/solo-ai/solo/pkg/types"
)

func TestGraph_TopologicalOrder(t *testing.T) {
	wf := &types.Workflow{
		ID:   "test",
		Name: "Test",
		Nodes: []types.WorkflowNode{
			{ID: "a", Name: "A", Executor: "x", Inputs: []string{}},
			{ID: "b", Name: "B", Executor: "x", Inputs: []string{"a"}},
			{ID: "c", Name: "C", Executor: "x", Inputs: []string{"a"}},
			{ID: "d", Name: "D", Executor: "x", Inputs: []string{"b", "c"}},
		},
		EntryPoint: "a",
	}

	g, err := NewGraph(wf)
	if err != nil {
		t.Fatal(err)
	}

	order, err := g.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}

	// 'a' must come first (no deps), 'd' last (depends on b, c)
	if order[0] != "a" {
		t.Errorf("expected first node 'a', got %s", order[0])
	}
	if order[len(order)-1] != "d" {
		t.Errorf("expected last node 'd', got %s", order[len(order)-1])
	}
}
