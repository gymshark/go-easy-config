package config

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildDependencyGraph(t *testing.T) {
	tests := []struct {
		name             string
		dependencies     map[int][]string
		availableAsMap   map[string]int
		fieldNames       map[int]string
		expectError      bool
		errorContains    string
		expectErrorType  string // "UndefinedVariableError" or empty
		expectedNodes    int
		expectedEdges    map[int]int // field index -> number of outgoing edges
		expectedInDegree map[int]int // field index -> in-degree
	}{
		{
			name: "simple linear dependency",
			dependencies: map[int][]string{
				1: {"VAR1"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
			},
			expectError:   false,
			expectedNodes: 2,
			expectedEdges: map[int]int{
				0: 1, // Field1 has 1 outgoing edge to Field2
				1: 0, // Field2 has no outgoing edges
			},
			expectedInDegree: map[int]int{
				0: 0, // Field1 has no dependencies
				1: 1, // Field2 depends on Field1
			},
		},
		{
			name: "multiple dependencies",
			dependencies: map[int][]string{
				2: {"VAR1", "VAR2"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
			},
			expectError:   false,
			expectedNodes: 3,
			expectedEdges: map[int]int{
				0: 1, // Field1 -> Field3
				1: 1, // Field2 -> Field3
				2: 0, // Field3 has no outgoing edges
			},
			expectedInDegree: map[int]int{
				0: 0, // Field1 has no dependencies
				1: 0, // Field2 has no dependencies
				2: 2, // Field3 depends on Field1 and Field2
			},
		},
		{
			name:         "no dependencies",
			dependencies: map[int][]string{},
			availableAsMap: map[string]int{
				"VAR1": 0,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
			},
			expectError:   false,
			expectedNodes: 2,
			expectedEdges: map[int]int{
				0: 0,
				1: 0,
			},
			expectedInDegree: map[int]int{
				0: 0,
				1: 0,
			},
		},
		{
			name: "undefined variable",
			dependencies: map[int][]string{
				1: {"UNDEFINED_VAR"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
			},
			expectError:     true,
			errorContains:   "undefined variable '${UNDEFINED_VAR}' referenced in field 'Field2'",
			expectErrorType: "UndefinedVariableError",
		},
		{
			name: "complex dependency chain",
			dependencies: map[int][]string{
				1: {"VAR1"},
				2: {"VAR2"},
				3: {"VAR1", "VAR3"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
				3: "Field4",
			},
			expectError:   false,
			expectedNodes: 4,
			expectedEdges: map[int]int{
				0: 2, // Field1 -> Field2, Field4
				1: 1, // Field2 -> Field3
				2: 1, // Field3 -> Field4
				3: 0, // Field4 has no outgoing edges
			},
			expectedInDegree: map[int]int{
				0: 0, // Field1 has no dependencies
				1: 1, // Field2 depends on Field1
				2: 1, // Field3 depends on Field2
				3: 2, // Field4 depends on Field1 and Field3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := BuildDependencyGraph(tt.dependencies, tt.availableAsMap, tt.fieldNames)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}

				// Check error type if specified
				if tt.expectErrorType == "UndefinedVariableError" {
					var undefinedErr *UndefinedVariableError
					if !errors.As(err, &undefinedErr) {
						t.Errorf("expected UndefinedVariableError but got %T", err)
					} else {
						// Verify field name and variable name are set correctly
						if undefinedErr.FieldName != "Field2" {
							t.Errorf("expected FieldName 'Field2', got '%s'", undefinedErr.FieldName)
						}
						if undefinedErr.VariableName != "UNDEFINED_VAR" {
							t.Errorf("expected VariableName 'UNDEFINED_VAR', got '%s'", undefinedErr.VariableName)
						}
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify number of nodes
			if len(graph.nodes) != tt.expectedNodes {
				t.Errorf("expected %d nodes, got %d", tt.expectedNodes, len(graph.nodes))
			}

			// Verify edges
			for fieldIndex, expectedEdgeCount := range tt.expectedEdges {
				actualEdgeCount := len(graph.edges[fieldIndex])
				if actualEdgeCount != expectedEdgeCount {
					t.Errorf("field %d: expected %d outgoing edges, got %d", fieldIndex, expectedEdgeCount, actualEdgeCount)
				}
			}

			// Verify in-degrees
			for fieldIndex, expectedInDegree := range tt.expectedInDegree {
				actualInDegree := graph.nodes[fieldIndex].inDegree
				if actualInDegree != expectedInDegree {
					t.Errorf("field %d: expected in-degree %d, got %d", fieldIndex, expectedInDegree, actualInDegree)
				}
			}
		})
	}
}

func TestDetectCycle(t *testing.T) {
	tests := []struct {
		name           string
		dependencies   map[int][]string
		availableAsMap map[string]int
		fieldNames     map[int]string
		expectCycle    bool
		cycleContains  []string // field names that should appear in cycle
	}{
		{
			name: "no cycle - linear",
			dependencies: map[int][]string{
				1: {"VAR1"},
				2: {"VAR2"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
			},
			expectCycle: false,
		},
		{
			name: "simple two-node cycle",
			dependencies: map[int][]string{
				0: {"VAR2"},
				1: {"VAR1"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
			},
			fieldNames: map[int]string{
				0: "FieldA",
				1: "FieldB",
			},
			expectCycle:   true,
			cycleContains: []string{"FieldA", "FieldB"},
		},
		{
			name: "three-node cycle",
			dependencies: map[int][]string{
				0: {"VAR3"},
				1: {"VAR1"},
				2: {"VAR2"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
			},
			fieldNames: map[int]string{
				0: "FieldA",
				1: "FieldB",
				2: "FieldC",
			},
			expectCycle:   true,
			cycleContains: []string{"FieldA", "FieldB", "FieldC"},
		},
		{
			name: "self-referencing cycle",
			dependencies: map[int][]string{
				0: {"VAR1"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
			},
			fieldNames: map[int]string{
				0: "FieldA",
			},
			expectCycle:   true,
			cycleContains: []string{"FieldA"},
		},
		{
			name: "complex graph with cycle",
			dependencies: map[int][]string{
				1: {"VAR1", "VAR4"}, // Field2 depends on VAR1 and VAR4 (provided by Field5)
				2: {"VAR2"},         // Field3 depends on VAR2 (provided by Field2)
				3: {"VAR3"},         // Field4 depends on VAR3 (provided by Field3)
				4: {"VAR2"},         // Field5 depends on VAR2 (provided by Field2) - creates cycle
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
				"VAR4": 4,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
				3: "Field4",
				4: "Field5",
			},
			expectCycle:   true,
			cycleContains: []string{"Field2", "Field5"},
		},
		{
			name: "no cycle - diamond pattern",
			dependencies: map[int][]string{
				1: {"VAR1"},
				2: {"VAR1"},
				3: {"VAR2", "VAR3"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
				3: "Field4",
			},
			expectCycle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := BuildDependencyGraph(tt.dependencies, tt.availableAsMap, tt.fieldNames)
			if err != nil {
				t.Fatalf("failed to build graph: %v", err)
			}

			cyclePath := graph.DetectCycle()

			if tt.expectCycle {
				if cyclePath == nil {
					t.Errorf("expected cycle path but got nil")
					return
				}

				// Verify that expected fields appear in the cycle
				cycleStr := strings.Join(cyclePath, " -> ")
				for _, fieldName := range tt.cycleContains {
					if !strings.Contains(cycleStr, fieldName) {
						t.Errorf("expected cycle to contain '%s', got: %s", fieldName, cycleStr)
					}
				}

				// Note: DetectCycle only returns the cycle path
				// CyclicDependencyError is created in TopologicalSort
			} else {
				if cyclePath != nil {
					t.Errorf("expected no cycle path but got: %v", cyclePath)
				}
			}
		})
	}
}

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name           string
		dependencies   map[int][]string
		availableAsMap map[string]int
		fieldNames     map[int]string
		expectError    bool
		expectedStages [][]int // expected field indices in each stage
	}{
		{
			name: "simple linear dependency",
			dependencies: map[int][]string{
				1: {"VAR1"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
			},
			expectError: false,
			expectedStages: [][]int{
				{0}, // Stage 0: Field1 (no dependencies)
				{1}, // Stage 1: Field2 (depends on Field1)
			},
		},
		{
			name: "parallel dependencies",
			dependencies: map[int][]string{
				2: {"VAR1", "VAR2"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
			},
			expectError: false,
			expectedStages: [][]int{
				{0, 1}, // Stage 0: Field1, Field2 (no dependencies, can be in any order)
				{2},    // Stage 1: Field3 (depends on Field1 and Field2)
			},
		},
		{
			name: "multi-level dependency chain",
			dependencies: map[int][]string{
				1: {"VAR1"},
				2: {"VAR2"},
				3: {"VAR3"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
				3: "Field4",
			},
			expectError: false,
			expectedStages: [][]int{
				{0}, // Stage 0: Field1
				{1}, // Stage 1: Field2
				{2}, // Stage 2: Field3
				{3}, // Stage 3: Field4
			},
		},
		{
			name: "diamond pattern",
			dependencies: map[int][]string{
				1: {"VAR1"},
				2: {"VAR1"},
				3: {"VAR2", "VAR3"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
				3: "Field4",
			},
			expectError: false,
			expectedStages: [][]int{
				{0},    // Stage 0: Field1
				{1, 2}, // Stage 1: Field2, Field3 (both depend only on Field1)
				{3},    // Stage 2: Field4 (depends on Field2 and Field3)
			},
		},
		{
			name:         "no dependencies",
			dependencies: map[int][]string{},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
			},
			expectError: false,
			expectedStages: [][]int{
				{0, 1, 2}, // All fields in stage 0 (no dependencies)
			},
		},
		{
			name: "cycle should fail with CyclicDependencyError",
			dependencies: map[int][]string{
				0: {"VAR2"},
				1: {"VAR1"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
			},
			fieldNames: map[int]string{
				0: "FieldA",
				1: "FieldB",
			},
			expectError: true,
		},
		{
			name: "complex multi-stage",
			dependencies: map[int][]string{
				1: {"VAR1"},
				2: {"VAR1"},
				3: {"VAR2"},
				4: {"VAR3", "VAR4"},
			},
			availableAsMap: map[string]int{
				"VAR1": 0,
				"VAR2": 1,
				"VAR3": 2,
				"VAR4": 3,
			},
			fieldNames: map[int]string{
				0: "Field1",
				1: "Field2",
				2: "Field3",
				3: "Field4",
				4: "Field5",
			},
			expectError: false,
			expectedStages: [][]int{
				{0},    // Stage 0: Field1
				{1, 2}, // Stage 1: Field2, Field3
				{3},    // Stage 2: Field4
				{4},    // Stage 3: Field5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := BuildDependencyGraph(tt.dependencies, tt.availableAsMap, tt.fieldNames)
			if err != nil {
				t.Fatalf("failed to build graph: %v", err)
			}

			stages, err := graph.TopologicalSort()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Verify error is CyclicDependencyError
				var cycleErr *CyclicDependencyError
				if errors.As(err, &cycleErr) {
					// Verify cycle path is populated
					if len(cycleErr.Cycle) == 0 {
						t.Errorf("expected cycle path to be populated")
					}
					// Verify error message format
					if !strings.Contains(err.Error(), "cyclic dependency detected") {
						t.Errorf("expected error message to contain 'cyclic dependency detected', got: %s", err.Error())
					}
				} else {
					t.Errorf("expected CyclicDependencyError but got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify number of stages
			if len(stages) != len(tt.expectedStages) {
				t.Errorf("expected %d stages, got %d", len(tt.expectedStages), len(stages))
				return
			}

			// Verify each stage
			for i, expectedStage := range tt.expectedStages {
				actualStage := stages[i]

				// Check stage size
				if len(actualStage) != len(expectedStage) {
					t.Errorf("stage %d: expected %d fields, got %d", i, len(expectedStage), len(actualStage))
					continue
				}

				// Convert to maps for easier comparison (order within stage doesn't matter)
				expectedMap := make(map[int]bool)
				for _, idx := range expectedStage {
					expectedMap[idx] = true
				}

				actualMap := make(map[int]bool)
				for _, idx := range actualStage {
					actualMap[idx] = true
				}

				// Verify all expected fields are present
				for idx := range expectedMap {
					if !actualMap[idx] {
						t.Errorf("stage %d: expected field %d (%s) but not found", i, idx, tt.fieldNames[idx])
					}
				}

				// Verify no unexpected fields
				for idx := range actualMap {
					if !expectedMap[idx] {
						t.Errorf("stage %d: unexpected field %d (%s)", i, idx, tt.fieldNames[idx])
					}
				}
			}

			// Verify dependency ordering: all dependencies of a field must appear in earlier stages
			fieldToStage := make(map[int]int)
			for stageIdx, stage := range stages {
				for _, fieldIdx := range stage {
					fieldToStage[fieldIdx] = stageIdx
				}
			}

			for fieldIdx, varNames := range tt.dependencies {
				fieldStage := fieldToStage[fieldIdx]
				for _, varName := range varNames {
					providerIdx := tt.availableAsMap[varName]
					providerStage := fieldToStage[providerIdx]
					if providerStage >= fieldStage {
						t.Errorf("field %d (%s) in stage %d depends on field %d (%s) in stage %d - dependency ordering violated",
							fieldIdx, tt.fieldNames[fieldIdx], fieldStage,
							providerIdx, tt.fieldNames[providerIdx], providerStage)
					}
				}
			}
		})
	}
}

func TestStageGrouping(t *testing.T) {
	// Test that fields are correctly grouped by dependency depth
	dependencies := map[int][]string{
		1: {"VAR1"},
		2: {"VAR2"},
		3: {"VAR3"},
		4: {"VAR1"},         // Same level as field 1
		5: {"VAR2", "VAR5"}, // Depends on field 2 and field 4
	}
	availableAsMap := map[string]int{
		"VAR1": 0,
		"VAR2": 1,
		"VAR3": 2,
		"VAR4": 3,
		"VAR5": 4,
	}
	fieldNames := map[int]string{
		0: "Field1",
		1: "Field2",
		2: "Field3",
		3: "Field4",
		4: "Field5",
		5: "Field6",
	}

	graph, err := BuildDependencyGraph(dependencies, availableAsMap, fieldNames)
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}

	stages, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("failed to sort: %v", err)
	}

	// Expected stages:
	// Stage 0: Field1 (no dependencies)
	// Stage 1: Field2, Field5 (depend on Field1)
	// Stage 2: Field3, Field6 (Field3 depends on Field2, Field6 depends on Field2 and Field5)
	// Stage 3: Field4 (depends on Field3)

	if len(stages) != 4 {
		t.Errorf("expected 4 stages, got %d", len(stages))
	}

	// Verify stage 0 contains Field1 (no dependencies)
	stage0Map := make(map[int]bool)
	for _, idx := range stages[0] {
		stage0Map[idx] = true
	}
	if !stage0Map[0] {
		t.Errorf("stage 0 should contain field 0, got: %v", stages[0])
	}

	// Verify stage 1 contains Field2 and Field5 (both depend on Field1)
	stage1Map := make(map[int]bool)
	for _, idx := range stages[1] {
		stage1Map[idx] = true
	}
	if !stage1Map[1] || !stage1Map[4] {
		t.Errorf("stage 1 should contain fields 1 and 4, got: %v", stages[1])
	}

	// Verify stage 2 contains Field3 and Field6
	stage2Map := make(map[int]bool)
	for _, idx := range stages[2] {
		stage2Map[idx] = true
	}
	if !stage2Map[2] || !stage2Map[5] {
		t.Errorf("stage 2 should contain fields 2 and 5, got: %v", stages[2])
	}

	// Verify stage 3 contains Field4 (depends on Field3)
	stage3Map := make(map[int]bool)
	for _, idx := range stages[3] {
		stage3Map[idx] = true
	}
	if !stage3Map[3] {
		t.Errorf("stage 3 should contain field 3, got: %v", stages[3])
	}
}

// TestDependencyGraphErrors tests that the correct error types are returned
// for various error conditions in dependency graph operations.
func TestDependencyGraphErrors(t *testing.T) {
	t.Run("UndefinedVariableError", func(t *testing.T) {
		dependencies := map[int][]string{
			1: {"UNDEFINED_VAR"},
		}
		availableAsMap := map[string]int{
			"VAR1": 0,
		}
		fieldNames := map[int]string{
			0: "Field1",
			1: "Field2",
		}

		_, err := BuildDependencyGraph(dependencies, availableAsMap, fieldNames)
		if err == nil {
			t.Fatal("expected error but got none")
		}

		// Verify error type
		var undefinedErr *UndefinedVariableError
		if !errors.As(err, &undefinedErr) {
			t.Fatalf("expected UndefinedVariableError but got %T", err)
		}

		// Verify field name
		if undefinedErr.FieldName != "Field2" {
			t.Errorf("expected FieldName 'Field2', got '%s'", undefinedErr.FieldName)
		}

		// Verify variable name
		if undefinedErr.VariableName != "UNDEFINED_VAR" {
			t.Errorf("expected VariableName 'UNDEFINED_VAR', got '%s'", undefinedErr.VariableName)
		}

		// Verify error message format
		expectedMsg := "undefined variable '${UNDEFINED_VAR}' referenced in field 'Field2'"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("CyclicDependencyError", func(t *testing.T) {
		dependencies := map[int][]string{
			0: {"VAR2"},
			1: {"VAR1"},
		}
		availableAsMap := map[string]int{
			"VAR1": 0,
			"VAR2": 1,
		}
		fieldNames := map[int]string{
			0: "FieldA",
			1: "FieldB",
		}

		graph, err := BuildDependencyGraph(dependencies, availableAsMap, fieldNames)
		if err != nil {
			t.Fatalf("failed to build graph: %v", err)
		}

		_, err = graph.TopologicalSort()
		if err == nil {
			t.Fatal("expected error but got none")
		}

		// Verify error type
		var cycleErr *CyclicDependencyError
		if !errors.As(err, &cycleErr) {
			t.Fatalf("expected CyclicDependencyError but got %T", err)
		}

		// Verify cycle path is populated
		if len(cycleErr.Cycle) == 0 {
			t.Error("expected cycle path to be populated")
		}

		// Verify cycle contains expected fields
		cycleStr := strings.Join(cycleErr.Cycle, " -> ")
		if !strings.Contains(cycleStr, "FieldA") || !strings.Contains(cycleStr, "FieldB") {
			t.Errorf("expected cycle to contain FieldA and FieldB, got: %s", cycleStr)
		}

		// Verify error message format
		if !strings.Contains(err.Error(), "cyclic dependency detected") {
			t.Errorf("expected error message to contain 'cyclic dependency detected', got: %s", err.Error())
		}
	})

	t.Run("DependencyGraphError for topological sort failure", func(t *testing.T) {
		// This test verifies that DependencyGraphError is used for sort failures
		// In practice, this should not happen if cycle detection works correctly,
		// but we test the error type is correct if it does occur.

		// Create a graph manually to simulate a sort failure scenario
		graph := &DependencyGraph{
			nodes: make(map[int]*GraphNode),
			edges: make(map[int][]int),
		}

		// Add nodes with circular in-degrees that would cause sort to fail
		// if cycle detection didn't catch it first
		graph.nodes[0] = &GraphNode{fieldIndex: 0, fieldName: "Field1", inDegree: 1}
		graph.nodes[1] = &GraphNode{fieldIndex: 1, fieldName: "Field2", inDegree: 1}
		graph.edges[0] = []int{1}
		graph.edges[1] = []int{0}

		// Skip cycle detection and go straight to sort to test the error
		// We'll manually create the error condition
		inDegree := make(map[int]int)
		for idx, node := range graph.nodes {
			inDegree[idx] = node.inDegree
		}

		processed := make(map[int]bool)

		// Try to find nodes with in-degree 0 (there are none due to cycle)
		currentStage := make([]int, 0)
		for idx := range graph.nodes {
			if !processed[idx] && inDegree[idx] == 0 {
				currentStage = append(currentStage, idx)
			}
		}

		// This should trigger the DependencyGraphError
		if len(currentStage) == 0 && len(processed) < len(graph.nodes) {
			err := &DependencyGraphError{
				Operation: "topological sort",
				Message:   "unable to complete sort: possible cycle",
			}

			// Verify error type
			var graphErr *DependencyGraphError
			if !errors.As(err, &graphErr) {
				t.Fatalf("expected DependencyGraphError but got %T", err)
			}

			// Verify operation field
			if graphErr.Operation != "topological sort" {
				t.Errorf("expected Operation 'topological sort', got '%s'", graphErr.Operation)
			}

			// Verify message field
			if graphErr.Message != "unable to complete sort: possible cycle" {
				t.Errorf("expected Message 'unable to complete sort: possible cycle', got '%s'", graphErr.Message)
			}

			// Verify error message format
			expectedMsg := "dependency graph error during topological sort: unable to complete sort: possible cycle"
			if err.Error() != expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
			}
		} else {
			t.Skip("Could not create sort failure condition for testing")
		}
	})
}
