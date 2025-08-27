package tree

import (
	"os"
	"testing"
)

func TestNodeIndexingAfterDeletion(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tree_deletion_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tree
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	t.Log("=== TreeKEM 삭제 후 노드 번호 재할당 테스트 ===")

	// Add multiple users
	t.Log("\nStep 1: 여러 사용자 추가")
	users := []string{"alice", "bob", "charlie", "david"}
	for _, user := range users {
		err = tree.Insert(user, []byte(user+"_public_key"))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", user, err)
		}
	}

	t.Log("초기 트리 구조:")
	printStructure(t, tree)

	// Delete a leaf node
	t.Log("\nStep 2: 리프 노드 삭제 (alice)")
	err = tree.Delete("alice")
	if err != nil {
		t.Fatalf("Failed to delete alice: %v", err)
	}

	t.Log("alice 삭제 후:")
	printStructure(t, tree)

	// Delete another node
	t.Log("\nStep 3: 다른 노드 삭제 (bob)")
	err = tree.Delete("bob")
	if err != nil {
		t.Fatalf("Failed to delete bob: %v", err)
	}

	t.Log("bob 삭제 후:")
	printStructure(t, tree)

	// Verify node indices are consecutive and start from 0
	t.Log("\nStep 4: 노드 번호 연속성 검증")
	structure := tree.GetTreeStructure()
	nodeIndices := make([]int, 0, len(structure))
	for _, info := range structure {
		nodeIndices = append(nodeIndices, info.NodeIndex)
	}

	// Check if indices start from 0 and are consecutive
	expectedIndex := 0
	for i := 0; i < len(structure); i++ {
		found := false
		for _, index := range nodeIndices {
			if index == expectedIndex {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("노드 번호 %d가 누락됨", expectedIndex)
		}
		expectedIndex++
	}

	t.Log("✓ 노드 번호가 올바르게 재할당됨")

	// Test path calculations still work
	t.Log("\nStep 5: 경로 계산 검증")
	for name, info := range structure {
		if info.NodeType == "leaf" {
			path := calculatePathToRoot(tree, info.NodeIndex)
			t.Logf("%s에서 루트까지의 경로: %v", name, path)
			
			// Verify all nodes in path exist
			for _, nodeIndex := range path {
				node := tree.GetNodeByIndex(nodeIndex)
				if node == nil {
					t.Errorf("경로의 노드 %d를 찾을 수 없음", nodeIndex)
				}
			}
		}
	}
}

func TestMultipleDeletions(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tree_multiple_deletion_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tree
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	t.Log("=== 연속 삭제 테스트 ===")

	// Add 5 users
	users := []string{"alice", "bob", "charlie", "david", "eve"}
	for _, user := range users {
		err = tree.Insert(user, []byte(user+"_public_key"))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", user, err)
		}
	}

	t.Log("초기 상태:")
	printStructure(t, tree)

	// Delete users one by one
	deleteOrder := []string{"charlie", "bob", "eve"}
	for i, user := range deleteOrder {
		t.Logf("\nStep %d: %s 삭제", i+1, user)
		err = tree.Delete(user)
		if err != nil {
			t.Fatalf("Failed to delete %s: %v", user, err)
		}
		printStructure(t, tree)
		
		// Verify tree consistency
		verifyTreeConsistency(t, tree)
	}
}

func printStructure(t *testing.T, tree *Tree) {
	structure := tree.GetTreeStructure()
	for name, info := range structure {
		t.Logf("  %s: 노드번호=%d, 부모=%d, 타입=%s", 
			name, info.NodeIndex, info.ParentIndex, info.NodeType)
	}
}

func verifyTreeConsistency(t *testing.T, tree *Tree) {
	structure := tree.GetTreeStructure()
	
	// Check that all parent-child relationships are valid
	for name, info := range structure {
		if info.ParentIndex != -1 {
			// Find parent
			var parentFound bool
			for _, parentInfo := range structure {
				if parentInfo.NodeIndex == info.ParentIndex {
					parentFound = true
					break
				}
			}
			if !parentFound {
				t.Errorf("노드 %s의 부모 %d를 찾을 수 없음", name, info.ParentIndex)
			}
		}
		
		// Verify node can be found by index
		node := tree.GetNodeByIndex(info.NodeIndex)
		if node == nil {
			t.Errorf("노드 번호 %d로 노드를 찾을 수 없음", info.NodeIndex)
		}
	}
}