package config

// DependencyGraph represents a directed acyclic graph (DAG) of field dependencies.
// It is used to determine the order in which fields should be loaded to satisfy
// variable interpolation requirements.
type DependencyGraph struct {
	nodes map[int]*GraphNode
	edges map[int][]int // adjacency list: field index -> list of dependent field indices
}

// GraphNode represents a node in the dependency graph.
type GraphNode struct {
	fieldIndex int
	fieldName  string
	inDegree   int // number of incoming edges (dependencies)
}

// BuildDependencyGraph creates a directed acyclic graph from field dependencies.
// It takes a map of field indices to their variable dependencies and a map of
// variable names to field indices.
//
// Parameters:
//   - dependencies: map[fieldIndex][]variableName - which variables each field depends on
//   - availableAsMap: map[variableName]fieldIndex - which field provides each variable
//   - fieldNames: map[fieldIndex]fieldName - field names for error messages
//
// Returns:
//   - *DependencyGraph: the constructed graph
//   - error: if undefined variables are referenced
func BuildDependencyGraph(dependencies map[int][]string, availableAsMap map[string]int, fieldNames map[int]string) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		nodes: make(map[int]*GraphNode),
		edges: make(map[int][]int),
	}

	// Create nodes for all fields
	allFields := make(map[int]bool)
	for fieldIndex := range fieldNames {
		allFields[fieldIndex] = true
		graph.nodes[fieldIndex] = &GraphNode{
			fieldIndex: fieldIndex,
			fieldName:  fieldNames[fieldIndex],
			inDegree:   0,
		}
	}

	// Build edges based on dependencies
	for fieldIndex, varNames := range dependencies {
		for _, varName := range varNames {
			// Find which field provides this variable
			providerIndex, exists := availableAsMap[varName]
			if !exists {
				return nil, &UndefinedVariableError{
					FieldName:    fieldNames[fieldIndex],
					VariableName: varName,
				}
			}

			// Add edge from provider to dependent field
			graph.edges[providerIndex] = append(graph.edges[providerIndex], fieldIndex)
			graph.nodes[fieldIndex].inDegree++
		}
	}

	return graph, nil
}

// DetectCycle identifies circular dependencies in the graph using depth-first search.
// It returns the cycle path if found, or nil if the graph is acyclic.
//
// Returns:
//   - []string: field names in the cycle (e.g., ["FieldA", "FieldB", "FieldA"]), or nil if no cycle
func (g *DependencyGraph) DetectCycle() []string {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[int]int)
	path := make([]int, 0)

	var dfs func(int) bool
	dfs = func(nodeIndex int) bool {
		state[nodeIndex] = visiting
		path = append(path, nodeIndex)

		// Check all neighbors
		for _, neighbor := range g.edges[nodeIndex] {
			if state[neighbor] == visiting {
				// Found a back edge - cycle detected
				// Find where the cycle starts in the path
				cycleStart := len(path)
				for i := len(path) - 1; i >= 0; i-- {
					if path[i] == neighbor {
						cycleStart = i
						break
					}
				}
				// Build cycle path
				cyclePath := make([]string, 0)
				for i := cycleStart; i < len(path); i++ {
					cyclePath = append(cyclePath, g.nodes[path[i]].fieldName)
				}
				// Close the cycle
				cyclePath = append(cyclePath, g.nodes[neighbor].fieldName)
				return true
			}
			if state[neighbor] == unvisited {
				if dfs(neighbor) {
					return true
				}
			}
		}

		state[nodeIndex] = visited
		path = path[:len(path)-1]
		return false
	}

	// Try DFS from each unvisited node
	for nodeIndex := range g.nodes {
		if state[nodeIndex] == unvisited {
			if dfs(nodeIndex) {
				// Reconstruct cycle from path
				cyclePath := make([]string, 0)
				for _, idx := range path {
					cyclePath = append(cyclePath, g.nodes[idx].fieldName)
				}
				if len(cyclePath) > 0 {
					// Close the cycle
					cyclePath = append(cyclePath, cyclePath[0])
				}
				return cyclePath
			}
		}
	}

	return nil
}

// TopologicalSort performs a topological sort using Kahn's algorithm.
// It returns fields grouped by dependency level (stages).
//
// Stage 0 contains fields with no dependencies.
// Stage 1 contains fields that depend only on Stage 0 fields.
// Stage N contains fields that depend on fields from stages 0 to N-1.
//
// Returns:
//   - [][]int: fields grouped by dependency stage
//   - error: if a cycle is detected
func (g *DependencyGraph) TopologicalSort() ([][]int, error) {
	// First check for cycles
	if cyclePath := g.DetectCycle(); cyclePath != nil {
		return nil, &CyclicDependencyError{
			Cycle: cyclePath,
		}
	}

	// Create a copy of in-degrees to avoid modifying the graph
	inDegree := make(map[int]int)
	for idx, node := range g.nodes {
		inDegree[idx] = node.inDegree
	}

	stages := make([][]int, 0)
	processed := make(map[int]bool)

	// Process nodes level by level
	for len(processed) < len(g.nodes) {
		// Find all nodes with in-degree 0 (no remaining dependencies)
		currentStage := make([]int, 0)
		for idx := range g.nodes {
			if !processed[idx] && inDegree[idx] == 0 {
				currentStage = append(currentStage, idx)
			}
		}

		if len(currentStage) == 0 {
			// This shouldn't happen if cycle detection worked correctly
			return nil, &DependencyGraphError{
				Operation: "topological sort",
				Message:   "unable to complete sort: possible cycle",
			}
		}

		// Add current stage to result
		stages = append(stages, currentStage)

		// Mark nodes as processed and reduce in-degree of neighbors
		for _, idx := range currentStage {
			processed[idx] = true
			for _, neighbor := range g.edges[idx] {
				inDegree[neighbor]--
			}
		}
	}

	return stages, nil
}
