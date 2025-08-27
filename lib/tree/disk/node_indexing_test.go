package disk

import (
	"os"
	"testing"
)

func TestNodeIndexing(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tree_indexing_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tree
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	t.Log("=== TreeKEM 노드 번호 할당 테스트 ===")

	// Add first user
	t.Log("\nStep 1: Alice 추가")
	err = tree.Insert("alice", []byte("alice_public_key"))
	if err != nil {
		t.Fatalf("Failed to insert alice: %v", err)
	}

	structure := tree.GetTreeStructure()
	for name, info := range structure {
		t.Logf("  %s: 노드번호=%d, 부모번호=%d", name, info.NodeIndex, info.ParentIndex)
	}

	// Add second user
	t.Log("\nStep 2: Bob 추가")
	err = tree.Insert("bob", []byte("bob_public_key"))
	if err != nil {
		t.Fatalf("Failed to insert bob: %v", err)
	}

	structure = tree.GetTreeStructure()
	for name, info := range structure {
		t.Logf("  %s: 노드번호=%d, 부모번호=%d", name, info.NodeIndex, info.ParentIndex)
	}

	// Add third user
	t.Log("\nStep 3: Charlie 추가")
	err = tree.Insert("charlie", []byte("charlie_public_key"))
	if err != nil {
		t.Fatalf("Failed to insert charlie: %v", err)
	}

	structure = tree.GetTreeStructure()
	t.Log("\n최종 트리 구조:")
	for name, info := range structure {
		t.Logf("  %s: 노드번호=%d, 부모번호=%d, 타입=%s", name, info.NodeIndex, info.ParentIndex, info.NodeType)
	}

	// Test node relationship functions
	t.Log("\n=== 노드 관계 함수 테스트 ===")
	for name, info := range structure {
		node := tree.GetNodeByIndex(info.NodeIndex)
		if node == nil {
			t.Errorf("노드 %d를 찾을 수 없음", info.NodeIndex)
			continue
		}
		
		t.Logf("노드 %s (번호=%d):", name, info.NodeIndex)
		t.Logf("  부모 번호: %d", node.ParentIndex())
		t.Logf("  왼쪽 자식 번호: %d", node.LeftChildIndex())
		t.Logf("  오른쪽 자식 번호: %d", node.RightChildIndex())
		t.Logf("  형제 번호: %d", node.SiblingIndex())
		t.Logf("  왼쪽 자식인가? %v", node.IsLeftChild())
		t.Logf("  오른쪽 자식인가? %v", node.IsRightChild())
	}

	// Test TreeKEM path calculation
	t.Log("\n=== TreeKEM 경로 계산 테스트 ===")
	// 각 리프에서 루트까지의 경로를 출력
	for name, info := range structure {
		if info.NodeType == "leaf" {
			path := calculatePathToRoot(tree, info.NodeIndex)
			t.Logf("%s에서 루트까지의 경로: %v", name, path)
		}
	}
}

// calculatePathToRoot calculates the path from a leaf to root (including intermediate nodes)
func calculatePathToRoot(tree *Tree, leafIndex int) []int {
	var path []int
	current := tree.GetNodeByIndex(leafIndex)
	
	for current != nil {
		path = append(path, current.NodeIndex())
		parentIndex := current.ParentIndex()
		if parentIndex == -1 {
			break
		}
		current = tree.GetNodeByIndex(parentIndex)
	}
	
	return path
}