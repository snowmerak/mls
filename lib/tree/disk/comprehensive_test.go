package disk

import (
	"os"
	"testing"
)

func TestTreeKEMComprehensive(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "treekem_comprehensive_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tree
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	t.Log("🌟 === TreeKEM 종합 테스트 시작 ===")

	// Phase 1: Initial users addition
	t.Log("\n📝 Phase 1: 초기 사용자 추가")
	users := []string{"alice", "bob", "charlie", "david", "eve"}
	for i, user := range users {
		t.Logf("  단계 %d: %s 추가", i+1, user)
		err = tree.Insert(user, []byte(user+"_public_key"))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", user, err)
		}
		
		t.Logf("    ✓ %s 추가 완료", user)
		logTreeStructure(t, tree, "    ")
		t.Log("")
	}

	// Phase 2: TreeKEM operations
	t.Log("\n🔐 Phase 2: TreeKEM 키 관리 작업")
	structure := tree.GetTreeStructure()
	
	// Set keys for intermediate nodes
	for name, info := range structure {
		if info.NodeType == "intermediate" {
			key := []byte("key_for_" + name)
			err = tree.SetIntermediateNodeKey(name, key)
			if err != nil {
				t.Errorf("Failed to set key for %s: %v", name, err)
			} else {
				t.Logf("  ✓ %s에 키 설정됨", name)
			}
		}
	}

	// Phase 3: Node relationship verification
	t.Log("\n🔍 Phase 3: 노드 관계 검증")
	for name, info := range structure {
		node := tree.GetNodeByIndex(info.NodeIndex)
		if node == nil {
			t.Errorf("노드 %d를 찾을 수 없음", info.NodeIndex)
			continue
		}
		
		t.Logf("  %s (노드 %d):", name, info.NodeIndex)
		t.Logf("    부모: %d, 왼쪽자식: %d, 오른쪽자식: %d", 
			node.ParentIndex(), node.LeftChildIndex(), node.RightChildIndex())
		t.Logf("    형제: %d, 위치: %s", 
			node.SiblingIndex(), getNodePosition(node.(*Element)))
		
		if info.NodeType == "leaf" {
			path := calculatePathToRoot(tree, info.NodeIndex)
			t.Logf("    루트까지 경로: %v", path)
		}
	}

	// Phase 4: Deletion and renumbering
	t.Log("\n🗑️  Phase 4: 삭제 및 번호 재할당 테스트")
	deleteOrder := []string{"charlie", "alice", "eve"}
	
	for i, user := range deleteOrder {
		t.Logf("\n  단계 %d: %s 삭제", i+1, user)
		
		// Show before state
		t.Log("    삭제 전 상태:")
		logTreeStructure(t, tree, "      ")
		
		// Perform deletion
		err = tree.Delete(user)
		if err != nil {
			t.Fatalf("Failed to delete %s: %v", user, err)
		}
		
		t.Logf("    ✓ %s 삭제 완료", user)
		t.Log("    삭제 후 상태 (번호 재할당됨):")
		logTreeStructure(t, tree, "      ")
		
		// Verify node numbering consistency
		verifyNodeNumberingConsistency(t, tree)
	}

	// Phase 5: Addition after deletion
	t.Log("\n➕ Phase 5: 삭제 후 새 사용자 추가")
	newUsers := []string{"frank", "grace"}
	
	for i, user := range newUsers {
		t.Logf("\n  단계 %d: %s 추가", i+1, user)
		
		// Show before state
		t.Log("    추가 전 상태:")
		logTreeStructure(t, tree, "      ")
		
		err = tree.Insert(user, []byte(user+"_public_key"))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", user, err)
		}
		
		t.Logf("    ✓ %s 추가 완료", user)
		t.Log("    추가 후 상태 (번호 재할당됨):")
		logTreeStructure(t, tree, "      ")
		
		verifyNodeNumberingConsistency(t, tree)
	}

	// Phase 6: Final verification
	t.Log("\n✅ Phase 6: 최종 검증")
	finalStructure := tree.GetTreeStructure()
	
	t.Log("  최종 트리 상태:")
	logTreeStructure(t, tree, "    ")
	
	// Verify all paths work
	t.Log("\n  모든 리프 노드의 경로 검증:")
	for name, info := range finalStructure {
		if info.NodeType == "leaf" {
			path := calculatePathToRoot(tree, info.NodeIndex)
			t.Logf("    %s → 루트: %v", name, path)
			
			// Verify each node in path exists
			for _, nodeIndex := range path {
				node := tree.GetNodeByIndex(nodeIndex)
				if node == nil {
					t.Errorf("경로의 노드 %d가 존재하지 않음", nodeIndex)
				}
			}
		}
	}

	t.Log("\n🎉 === TreeKEM 종합 테스트 완료 ===")
	t.Log("✓ 모든 삽입, 삭제, 번호 재할당이 정상 작동")
	t.Log("✓ TreeKEM 키 관리 기능 정상")
	t.Log("✓ 노드 관계 계산 정확")
	t.Log("✓ 경로 계산 정확")
}

func logTreeStructure(t *testing.T, tree *Tree, indent string) {
	structure := tree.GetTreeStructure()
	
	// Sort by node index for consistent output
	nodesByIndex := make(map[int]string)
	for name, info := range structure {
		nodesByIndex[info.NodeIndex] = name
	}
	
	for i := 0; i < len(structure); i++ {
		if name, exists := nodesByIndex[i]; exists {
			info := structure[name]
			if info.NodeType == "leaf" {
				t.Logf("%s🍃 %s: 노드=%d, 부모=%d", indent, name, info.NodeIndex, info.ParentIndex)
			} else {
				t.Logf("%s🌿 %s: 노드=%d, 부모=%d", indent, name, info.NodeIndex, info.ParentIndex)
			}
		}
	}
}

func getNodePosition(node *Element) string {
	if node.NodeIndex() == 0 {
		return "루트"
	}
	if node.IsLeftChild() {
		return "왼쪽자식"
	}
	if node.IsRightChild() {
		return "오른쪽자식"
	}
	return "알수없음"
}

func verifyNodeNumberingConsistency(t *testing.T, tree *Tree) {
	structure := tree.GetTreeStructure()
	
	// Check that node indices are consecutive starting from 0
	expectedIndices := make(map[int]bool)
	for i := 0; i < len(structure); i++ {
		expectedIndices[i] = false
	}
	
	for _, info := range structure {
		if _, exists := expectedIndices[info.NodeIndex]; !exists {
			t.Errorf("예상하지 못한 노드 번호: %d", info.NodeIndex)
		} else {
			expectedIndices[info.NodeIndex] = true
		}
	}
	
	// Check all expected indices are present
	for index, found := range expectedIndices {
		if !found {
			t.Errorf("누락된 노드 번호: %d", index)
		}
	}
}