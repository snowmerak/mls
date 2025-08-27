package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLargeScaleOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mls_largescale_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create new tree
	diskTree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new tree: %v", err)
	}

	t.Log("=== Large Scale Test Started ===")
	
	// Test 1: Insert 1000 nodes
	nodeCount := 1000
	t.Logf("Inserting %d nodes...", nodeCount)
	
	start := time.Now()
	for i := 0; i < nodeCount; i++ {
		nodeName := fmt.Sprintf("node_%04d", i)
		keyMaterial := []byte(fmt.Sprintf("encryption_key_for_%s", nodeName))
		
		err := diskTree.Insert(nodeName, keyMaterial)
		if err != nil {
			t.Fatalf("Failed to insert node %s: %v", nodeName, err)
		}
		
		// Log progress every 100 insertions
		if (i+1)%100 == 0 {
			t.Logf("Inserted %d/%d nodes", i+1, nodeCount)
		}
	}
	insertDuration := time.Since(start)
	t.Logf("✓ Insertion completed in %v (avg: %v per node)", insertDuration, insertDuration/time.Duration(nodeCount))

	// Test 2: Verify all nodes can be found
	t.Log("Verifying all nodes can be found...")
	start = time.Now()
	
	for i := 0; i < nodeCount; i++ {
		nodeName := fmt.Sprintf("node_%04d", i)
		expectedKey := []byte(fmt.Sprintf("encryption_key_for_%s", nodeName))
		
		element, found := diskTree.Find(nodeName)
		if !found {
			t.Fatalf("Node %s not found", nodeName)
		}
		
		if element.Name() != nodeName {
			t.Fatalf("Expected name %s, got %s", nodeName, element.Name())
		}
		
		if string(element.Value()) != string(expectedKey) {
			t.Fatalf("Key mismatch for node %s", nodeName)
		}
		
		// Log progress every 200 searches
		if (i+1)%200 == 0 {
			t.Logf("Verified %d/%d nodes", i+1, nodeCount)
		}
	}
	searchDuration := time.Since(start)
	t.Logf("✓ Search verification completed in %v (avg: %v per search)", searchDuration, searchDuration/time.Duration(nodeCount))

	// Test 3: Check file system state
	t.Log("Checking file system state...")
	files, err := filepath.Glob(filepath.Join(tempDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	
	if len(files) != nodeCount {
		t.Fatalf("Expected %d files, found %d", nodeCount, len(files))
	}
	t.Logf("✓ File system verification passed: %d files created", len(files))

	// Test 4: Tree structure analysis
	t.Log("Analyzing tree structure...")
	head := diskTree.Head()
	if head == nil {
		t.Fatal("Head is nil")
	}
	
	leftDepth := calculateDepth(head.LeftChild())
	rightDepth := calculateDepth(head.RightChild())
	totalLeftNodes := head.LeftCount()
	totalRightNodes := head.RightCount()
	
	t.Logf("Tree structure analysis:")
	t.Logf("  Root node: %s", head.Name())
	t.Logf("  Left subtree depth: %d, nodes: %d", leftDepth, totalLeftNodes)
	t.Logf("  Right subtree depth: %d, nodes: %d", rightDepth, totalRightNodes)
	t.Logf("  Tree balance ratio: %.2f", float64(totalLeftNodes)/float64(totalRightNodes+1))

	// Test 5: Delete half of the nodes (every other node)
	deleteCount := nodeCount / 2
	t.Logf("Deleting %d nodes...", deleteCount)
	
	start = time.Now()
	deletedNodes := 0
	for i := 0; i < nodeCount; i += 2 {
		nodeName := fmt.Sprintf("node_%04d", i)
		
		err := diskTree.Delete(nodeName)
		if err != nil {
			t.Fatalf("Failed to delete node %s: %v", nodeName, err)
		}
		deletedNodes++
		
		// Log progress every 50 deletions
		if deletedNodes%50 == 0 {
			t.Logf("Deleted %d/%d nodes", deletedNodes, deleteCount)
		}
	}
	deleteDuration := time.Since(start)
	t.Logf("✓ Deletion completed in %v (avg: %v per deletion)", deleteDuration, deleteDuration/time.Duration(deleteCount))

	// Test 6: Verify deleted nodes are gone and remaining nodes still exist
	t.Log("Verifying deletion and remaining nodes...")
	remainingCount := 0
	
	for i := 0; i < nodeCount; i++ {
		nodeName := fmt.Sprintf("node_%04d", i)
		element, found := diskTree.Find(nodeName)
		
		if i%2 == 0 {
			// Should be deleted
			if found {
				t.Fatalf("Node %s should have been deleted but was found", nodeName)
			}
		} else {
			// Should still exist
			if !found {
				t.Fatalf("Node %s should exist but was not found", nodeName)
			}
			if element.Name() != nodeName {
				t.Fatalf("Name mismatch for remaining node %s", nodeName)
			}
			remainingCount++
		}
	}
	t.Logf("✓ Deletion verification passed: %d nodes remaining", remainingCount)

	// Test 7: Check file system cleanup
	t.Log("Checking file system cleanup...")
	files, err = filepath.Glob(filepath.Join(tempDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list files after deletion: %v", err)
	}
	
	expectedFiles := nodeCount - deleteCount
	if len(files) != expectedFiles {
		t.Fatalf("Expected %d files after deletion, found %d", expectedFiles, len(files))
	}
	t.Logf("✓ File cleanup verification passed: %d files remaining", len(files))

	// Test 8: Performance stress test - rapid insertions and deletions
	t.Log("Running performance stress test...")
	stressNodes := 100
	
	start = time.Now()
	for round := 0; round < 5; round++ {
		// Insert stress nodes
		for i := 0; i < stressNodes; i++ {
			nodeName := fmt.Sprintf("stress_%d_%d", round, i)
			keyMaterial := []byte(fmt.Sprintf("stress_key_%d_%d", round, i))
			
			err := diskTree.Insert(nodeName, keyMaterial)
			if err != nil {
				t.Fatalf("Stress test insert failed: %v", err)
			}
		}
		
		// Delete stress nodes
		for i := 0; i < stressNodes; i++ {
			nodeName := fmt.Sprintf("stress_%d_%d", round, i)
			
			err := diskTree.Delete(nodeName)
			if err != nil {
				t.Fatalf("Stress test delete failed: %v", err)
			}
		}
		
		t.Logf("Stress test round %d completed", round+1)
	}
	stressDuration := time.Since(start)
	t.Logf("✓ Stress test completed in %v", stressDuration)

	// Final tree state analysis
	head = diskTree.Head()
	if head != nil {
		finalLeftCount := head.LeftCount()
		finalRightCount := head.RightCount()
		t.Logf("Final tree state: Left=%d, Right=%d, Total=%d", 
			finalLeftCount, finalRightCount, finalLeftCount+finalRightCount+1)
	}

	t.Log("=== Large Scale Test Completed Successfully ===")
}

// Helper function to calculate tree depth
func calculateDepth(element interface{}) int {
	if element == nil {
		return 0
	}
	
	if diskElement, ok := element.(*Element); ok && diskElement != nil {
		leftDepth := calculateDepth(diskElement.leftChild)
		rightDepth := calculateDepth(diskElement.rightChild)
		
		if leftDepth > rightDepth {
			return leftDepth + 1
		}
		return rightDepth + 1
	}
	
	return 0
}

func TestTreeKEMGroupScenario(t *testing.T) {
	// Simulate a TreeKEM group scenario
	tempDir, err := os.MkdirTemp("", "mls_treekem_group_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskTree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new tree: %v", err)
	}

	t.Log("=== TreeKEM Group Scenario Test ===")

	// Simulate group members joining (common TreeKEM scenario)
	groupMembers := []string{
		"alice@example.com", "bob@example.com", "charlie@example.com",
		"diana@example.com", "eve@example.com", "frank@example.com",
		"grace@example.com", "henry@example.com", "iris@example.com",
		"jack@example.com", "karen@example.com", "leo@example.com",
	}

	t.Logf("Simulating group with %d members joining...", len(groupMembers))

	// Phase 1: Members join one by one
	for i, member := range groupMembers {
		// In TreeKEM, each member would have associated key material
		keyMaterial := []byte(fmt.Sprintf("member_key_%s_%d", member, i))
		
		err := diskTree.Insert(member, keyMaterial)
		if err != nil {
			t.Fatalf("Failed to add member %s: %v", member, err)
		}
		
		t.Logf("Member %d joined: %s", i+1, member)
		
		// Log current tree state
		head := diskTree.Head()
		if head != nil {
			t.Logf("  Current tree size: Left=%d, Right=%d", 
				head.LeftCount(), head.RightCount())
		}
	}

	// Phase 2: Verify all members are in the group
	t.Log("Verifying all group members...")
	for _, member := range groupMembers {
		element, found := diskTree.Find(member)
		if !found {
			t.Fatalf("Group member %s not found", member)
		}
		t.Logf("  ✓ %s (key: %s)", member, string(element.Value()[:20])+"...")
	}

	// Phase 3: Simulate some members leaving (TreeKEM remove operations)
	leavingMembers := []string{"charlie@example.com", "eve@example.com", "henry@example.com"}
	
	t.Logf("Simulating %d members leaving...", len(leavingMembers))
	for _, member := range leavingMembers {
		err := diskTree.Delete(member)
		if err != nil {
			t.Fatalf("Failed to remove member %s: %v", member, err)
		}
		t.Logf("Member left: %s", member)
	}

	// Phase 4: Verify remaining members
	remainingMembers := make([]string, 0)
	for _, member := range groupMembers {
		if !contains(leavingMembers, member) {
			remainingMembers = append(remainingMembers, member)
		}
	}

	t.Logf("Verifying %d remaining members...", len(remainingMembers))
	for _, member := range remainingMembers {
		_, found := diskTree.Find(member)
		if !found {
			t.Fatalf("Remaining member %s not found", member)
		}
		t.Logf("  ✓ %s still in group", member)
	}

	// Phase 5: New members join
	newMembers := []string{"mike@example.com", "nancy@example.com", "oscar@example.com"}
	
	t.Logf("Adding %d new members...", len(newMembers))
	for i, member := range newMembers {
		keyMaterial := []byte(fmt.Sprintf("new_member_key_%s_%d", member, i))
		
		err := diskTree.Insert(member, keyMaterial)
		if err != nil {
			t.Fatalf("Failed to add new member %s: %v", member, err)
		}
		t.Logf("New member joined: %s", member)
	}

	// Final verification
	finalMemberCount := len(remainingMembers) + len(newMembers)
	head := diskTree.Head()
	actualCount := 1 + head.LeftCount() + head.RightCount()
	
	if actualCount != finalMemberCount {
		t.Fatalf("Expected %d total members, got %d", finalMemberCount, actualCount)
	}

	t.Logf("✓ Final group state: %d active members", finalMemberCount)
	t.Log("=== TreeKEM Group Scenario Test Completed ===")
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func BenchmarkDiskTreeOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "mls_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskTree, err := NewTree(tempDir)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}

	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		nodeName := fmt.Sprintf("bench_node_%04d", i)
		keyMaterial := []byte(fmt.Sprintf("benchmark_key_%d", i))
		diskTree.Insert(nodeName, keyMaterial)
	}

	b.ResetTimer()

	b.Run("Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeName := fmt.Sprintf("insert_bench_%d", i)
			keyMaterial := []byte(fmt.Sprintf("insert_key_%d", i))
			diskTree.Insert(nodeName, keyMaterial)
		}
	})

	b.Run("Find", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeName := fmt.Sprintf("bench_node_%04d", i%100)
			diskTree.Find(nodeName)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeName := fmt.Sprintf("insert_bench_%d", i)
			diskTree.Delete(nodeName)
		}
	})
}