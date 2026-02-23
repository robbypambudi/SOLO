package types

// WorkflowNode represents a node in the reasoning graph.
// Supports patterns: tool-augmented reasoning, multi-step planning, self-reflection.
type WorkflowNode struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Executor     string            `json:"executor"`      // e.g., "llm", "retriever", "planner", "agent"
	Config       map[string]string `json:"config,omitempty"`
	Inputs       []string          `json:"inputs"`        // IDs of nodes whose output feeds this node
	Conditional  string            `json:"conditional,omitempty"` // Optional: expression for conditional execution
}

// Workflow represents a directed graph of reasoning steps.
type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Nodes       []WorkflowNode `json:"nodes"`
	EntryPoint  string         `json:"entry_point"` // Node ID where execution starts
}
