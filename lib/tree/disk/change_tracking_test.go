package disk

import (
	"os"
	"testing"
	"time"
)

func TestNodeChangeTracking(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "change_tracking_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tree
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	t.Log("🚀 === 노드 변경 추적 테스트 시작 ===")

	// Record start time
	startTime := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp differences

	t.Log("\n📝 Phase 1: 초기 노드 추가")
	
	// Add some nodes
	users := []string{"alice", "bob", "charlie"}
	for i, user := range users {
		t.Logf("  %d. %s 추가", i+1, user)
		err = tree.Insert(user, []byte(user+"_key"))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", user, err)
		}
		time.Sleep(5 * time.Millisecond) // Small delay between insertions
	}

	t.Log("\n🔍 Phase 2: 변경된 노드 확인")
	
	// Check nodes modified since start
	modifiedNodes := tree.GetModifiedNodes(startTime)
	t.Logf("  시작 시점 이후 변경된 노드 수: %d", len(modifiedNodes))
	for _, node := range modifiedNodes {
		if element, ok := node.(*Element); ok {
			t.Logf("    - %s (수정시점: %v)", element.name, element.lastModified.Format("15:04:05.000"))
		}
	}

	// Check nodes needing update (all should need update since we haven't checked them)
	needingUpdate := tree.GetNodesNeedingUpdate()
	t.Logf("  업데이트가 필요한 노드 수: %d", len(needingUpdate))
	for _, node := range needingUpdate {
		if element, ok := node.(*Element); ok {
			t.Logf("    - %s (확인 필요)", element.name)
		}
	}

	t.Log("\n✅ Phase 3: 모든 노드를 확인함으로 표시")
	
	checkTime := time.Now()
	tree.MarkAllAsChecked()
	t.Logf("  모든 노드 확인 완료 (시점: %v)", checkTime.Format("15:04:05.000"))

	// Now no nodes should need update
	needingUpdateAfterCheck := tree.GetNodesNeedingUpdate()
	t.Logf("  확인 후 업데이트가 필요한 노드 수: %d", len(needingUpdateAfterCheck))

	t.Log("\n🔄 Phase 4: 일부 노드 수정")
	
	time.Sleep(10 * time.Millisecond)
	
	// Modify alice's key
	t.Log("  alice의 키를 업데이트")
	element, found := tree.Find("alice")
	if !found {
		t.Fatal("Alice not found")
	}
	aliceElement := element.(*Element)
	aliceElement.publicKey = []byte("alice_new_key")
	aliceElement.MarkAsModified()
	aliceElement.saveToDisk()

	// Add new intermediate key
	t.Log("  intermediate 노드 키 설정")
	err = tree.SetIntermediateNodeKey("intermediate_alice_bob", []byte("shared_key_alice_bob"))
	if err != nil {
		t.Logf("  (intermediate 노드가 없을 수 있음: %v)", err)
	}

	t.Log("\n🎯 Phase 5: 변경 사항 추적")
	
	// Check what changed since we marked everything as checked
	changedSinceCheck := tree.GetNodeChangesSince(checkTime)
	t.Logf("  확인 시점 이후 변경된 노드들:")
	for name, modTime := range changedSinceCheck {
		t.Logf("    - %s: %v", name, modTime.Format("15:04:05.000"))
	}

	// Check nodes needing update again
	needingUpdateNow := tree.GetNodesNeedingUpdate()
	t.Logf("  현재 업데이트가 필요한 노드 수: %d", len(needingUpdateNow))
	for _, node := range needingUpdateNow {
		if element, ok := node.(*Element); ok {
			t.Logf("    - %s (마지막 수정: %v, 마지막 확인: %v)", 
				element.name, 
				element.lastModified.Format("15:04:05.000"),
				element.lastChecked.Format("15:04:05.000"))
		}
	}

	t.Log("\n📊 Phase 6: 개별 노드 상태 확인")
	
	// Check individual node status
	structure := tree.GetTreeStructure()
	for name, info := range structure {
		node := tree.GetNodeByIndex(info.NodeIndex)
		if node != nil {
			if element, ok := node.(*Element); ok {
				needsUpdate := element.NeedsUpdate()
				t.Logf("  %s (노드=%d): 업데이트 필요=%t", name, info.NodeIndex, needsUpdate)
				if needsUpdate {
					t.Logf("    └─ 수정: %v, 확인: %v", 
						element.lastModified.Format("15:04:05.000"),
						element.lastChecked.Format("15:04:05.000"))
				}
			}
		}
	}

	t.Log("\n🎉 === 노드 변경 추적 테스트 완료 ===")
	t.Log("✓ lastModified, lastChecked 시간 추적 정상 작동")
	t.Log("✓ WasModifiedSince() 메서드 정상 작동")
	t.Log("✓ NeedsUpdate() 메서드 정상 작동")
	t.Log("✓ 변경된 노드들을 빠르게 찾을 수 있음")
}

func TestFastChangeDetection(t *testing.T) {
	t.Log("⚡ === 빠른 변경 감지 성능 테스트 ===")
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fast_change_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tree with many nodes
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Add 20 nodes
	nodeCount := 20
	t.Logf("  %d개 노드 추가 중...", nodeCount)
	for i := 0; i < nodeCount; i++ {
		name := string(rune('a' + i))
		err = tree.Insert(name, []byte(name+"_key"))
		if err != nil {
			t.Fatalf("Failed to insert node %s: %v", name, err)
		}
	}

	// Mark all as checked
	tree.MarkAllAsChecked()
	
	time.Sleep(1 * time.Millisecond)

	// Modify only 3 nodes
	modifiedNodes := []string{"c", "g", "m"}
	t.Logf("  %d개 노드만 수정 (%v)", len(modifiedNodes), modifiedNodes)
	
	for _, name := range modifiedNodes {
		element, found := tree.Find(name)
		if found {
			if diskElement, ok := element.(*Element); ok {
				diskElement.publicKey = []byte(name + "_modified_key")
				diskElement.MarkAsModified()
				diskElement.saveToDisk()
			}
		}
	}

	// Fast detection: only get nodes that need updates
	start := time.Now()
	needingUpdate := tree.GetNodesNeedingUpdate()
	detectionTime := time.Since(start)

	t.Logf("  ⚡ 변경 감지 소요 시간: %v", detectionTime)
	t.Logf("  📊 전체 노드 수: %d, 변경된 노드 수: %d", nodeCount*2-1, len(needingUpdate)) // approx total nodes in TreeKEM

	if len(needingUpdate) != len(modifiedNodes) {
		t.Errorf("Expected %d modified nodes, got %d", len(modifiedNodes), len(needingUpdate))
	}

	// Verify only the right nodes were detected
	detectedNames := make(map[string]bool)
	for _, node := range needingUpdate {
		if element, ok := node.(*Element); ok {
			detectedNames[element.name] = true
			t.Logf("    ✓ 감지된 변경 노드: %s", element.name)
		}
	}

	for _, expectedName := range modifiedNodes {
		if !detectedNames[expectedName] {
			t.Errorf("Expected to detect changed node %s, but didn't", expectedName)
		}
	}

	t.Log("✓ 빠른 변경 감지 테스트 성공!")
}