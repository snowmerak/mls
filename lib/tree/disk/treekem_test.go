package disk

import (
	"fmt"
	"os"
	"testing"
)

func TestTreeKEMStructure(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mls_treekem_structure_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskTree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new tree: %v", err)
	}

	t.Log("=== TreeKEM Structure Test ===")

	// Test 1: Single member (should be root)
	t.Log("Adding first member...")
	err = diskTree.Insert("alice", []byte("alice_key_material"))
	if err != nil {
		t.Fatalf("Failed to insert alice: %v", err)
	}

	head := diskTree.Head()
	if head.Name() != "alice" {
		t.Errorf("Expected head to be alice, got %s", head.Name())
	}
	
	if !head.(*Element).IsLeaf() {
		t.Error("Single member should be a leaf")
	}

	// Test 2: Second member (should create intermediate node)
	t.Log("Adding second member...")
	err = diskTree.Insert("bob", []byte("bob_key_material"))
	if err != nil {
		t.Fatalf("Failed to insert bob: %v", err)
	}

	head = diskTree.Head()
	t.Logf("Head after second member: %s", head.Name())
	
	// Head should now be an intermediate node with alice and bob as children
	if head.(*Element).IsLeaf() {
		t.Error("Head should not be a leaf after adding second member")
	}

	// Check children
	if head.LeftChild() == nil || head.RightChild() == nil {
		t.Error("Head should have both left and right children")
	}

	// Test 3: Get all leaves
	leaves := diskTree.GetLeaves()
	t.Logf("Found %d leaves", len(leaves))
	
	if len(leaves) != 2 {
		t.Errorf("Expected 2 leaves, got %d", len(leaves))
	}

	leafNames := make(map[string]bool)
	for _, leaf := range leaves {
		leafNames[leaf.Name()] = true
		t.Logf("Leaf: %s", leaf.Name())
	}

	if !leafNames["alice"] || !leafNames["bob"] {
		t.Error("Should have alice and bob as leaves")
	}

	// Test 4: Add more members
	members := []string{"charlie", "diana", "eve"}
	for _, member := range members {
		t.Logf("Adding member: %s", member)
		err = diskTree.Insert(member, []byte(fmt.Sprintf("%s_key_material", member)))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", member, err)
		}
		
		leaves = diskTree.GetLeaves()
		t.Logf("  Total leaves after adding %s: %d", member, len(leaves))
	}

	// Final verification
	finalLeaves := diskTree.GetLeaves()
	if len(finalLeaves) != 5 {
		t.Errorf("Expected 5 final leaves, got %d", len(finalLeaves))
	}

	expectedMembers := []string{"alice", "bob", "charlie", "diana", "eve"}
	finalLeafNames := make(map[string]bool)
	for _, leaf := range finalLeaves {
		finalLeafNames[leaf.Name()] = true
	}

	for _, expected := range expectedMembers {
		if !finalLeafNames[expected] {
			t.Errorf("Expected member %s not found in leaves", expected)
		}
	}

	t.Log("✓ TreeKEM structure test completed successfully")
}

func TestTreeKEMPath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mls_treekem_path_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskTree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new tree: %v", err)
	}

	t.Log("=== TreeKEM Path Test ===")

	// Add several members
	members := []string{"alice", "bob", "charlie", "diana"}
	for _, member := range members {
		err = diskTree.Insert(member, []byte(fmt.Sprintf("%s_key", member)))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", member, err)
		}
	}

	// Test path from each leaf to root
	for _, member := range members {
		t.Logf("Testing path for member: %s", member)
		
		path, err := diskTree.GetPath(member)
		if err != nil {
			t.Errorf("Failed to get path for %s: %v", member, err)
			continue
		}

		t.Logf("  Path length: %d", len(path))
		for i, node := range path {
			t.Logf("    [%d] %s (isLeaf: %v)", i, node.Name(), node.(*Element).IsLeaf())
		}

		// Verify path properties
		if len(path) == 0 {
			t.Errorf("Path for %s should not be empty", member)
			continue
		}

		// Last node in path should be the leaf we're looking for
		lastNode := path[len(path)-1]
		if lastNode.Name() != member {
			t.Errorf("Last node in path should be %s, got %s", member, lastNode.Name())
		}

		if !lastNode.(*Element).IsLeaf() {
			t.Errorf("Last node in path for %s should be a leaf", member)
		}

		// First node should be the root
		rootNode := path[0]
		if rootNode != diskTree.Head() {
			t.Errorf("First node in path should be the root")
		}
	}

	t.Log("✓ TreeKEM path test completed successfully")
}

func TestTreeKEMLeafOnlyInsertion(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mls_treekem_leaf_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskTree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new tree: %v", err)
	}

	t.Log("=== TreeKEM Leaf-Only Insertion Test ===")

	// Add 8 members to create a more complex tree
	members := []string{"alice", "bob", "charlie", "diana", "eve", "frank", "grace", "henry"}
	
	for i, member := range members {
		t.Logf("Adding member %d: %s", i+1, member)
		
		err = diskTree.Insert(member, []byte(fmt.Sprintf("%s_secret_key", member)))
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", member, err)
		}

		// Verify only the expected number of leaves exist
		leaves := diskTree.GetLeaves()
		expectedLeafCount := i + 1
		
		if len(leaves) != expectedLeafCount {
			t.Errorf("After adding %s, expected %d leaves, got %d", member, expectedLeafCount, len(leaves))
		}

		// Verify all leaves are actual members (not intermediate nodes)
		for _, leaf := range leaves {
			leafName := leaf.Name()
			isExpectedMember := false
			for j := 0; j <= i; j++ {
				if leafName == members[j] {
					isExpectedMember = true
					break
				}
			}
			
			if !isExpectedMember {
				t.Errorf("Unexpected leaf node: %s (should only have member nodes as leaves)", leafName)
			}
		}

		t.Logf("  ✓ %d leaves confirmed", len(leaves))
	}

	// Final structure verification
	finalLeaves := diskTree.GetLeaves()
	t.Logf("Final tree structure:")
	t.Logf("  Total leaves (members): %d", len(finalLeaves))
	
	for _, leaf := range finalLeaves {
		t.Logf("    Member: %s", leaf.Name())
	}

	// Verify tree depth is reasonable (should be log2(n) for balanced tree)
	maxDepth := 0
	for _, member := range members {
		path, err := diskTree.GetPath(member)
		if err != nil {
			t.Errorf("Failed to get path for %s: %v", member, err)
			continue
		}
		if len(path) > maxDepth {
			maxDepth = len(path)
		}
	}

	expectedMaxDepth := 4 // log2(8) = 3, plus 1 for inclusive path
	if maxDepth > expectedMaxDepth+1 { // Allow some flexibility
		t.Errorf("Tree depth %d seems too deep for %d members (expected around %d)", maxDepth, len(members), expectedMaxDepth)
	}

	t.Logf("  Max path depth: %d", maxDepth)
	t.Log("✓ TreeKEM leaf-only insertion test completed successfully")
}

// TestTreeKEMPublicKeyDerivation tests the TreeKEM public key derivation functionality
func TestTreeKEMPublicKeyDerivation(t *testing.T) {
	t.Log("=== TreeKEM Public Key Derivation Test ===")
	
	// Create test tree
	tempDir := t.TempDir()
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}
	
	// Simulate client public keys (in real TreeKEM, these come from key exchange)
	clientPublicKeys := map[string][]byte{
		"alice@example.com":   []byte("alice_public_key_x25519_abcd1234"),
		"bob@example.com":     []byte("bob_public_key_x25519_efgh5678"),
		"charlie@example.com": []byte("charlie_public_key_x25519_ijkl9012"),
	}
	
	t.Log("Adding members with their public keys...")
	
	for email, pubKey := range clientPublicKeys {
		err := tree.Insert(email, pubKey)
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", email, err)
		}
		
		// Verify group public key exists and is derived
		groupPubKey := tree.GetGroupPublicKey()
		if len(groupPubKey) == 0 {
			t.Errorf("Group public key should not be empty after adding %s", email)
		} else {
			t.Logf("  Group public key after adding %s: %x...", email, groupPubKey[:8])
		}
	}
	
	// Check TreeKEM properties
	leaves := tree.GetLeaves()
	t.Logf("Found %d leaf nodes (actual users)", len(leaves))
	
	for _, leaf := range leaves {
		element := leaf.(*Element)
		if element.nodeType != "leaf" {
			t.Errorf("Leaf node %s should have nodeType 'leaf'", element.name)
		}
		if len(element.publicKey) == 0 {
			t.Errorf("Leaf node %s should have a public key", element.name)
		}
	}
	
	t.Log("TreeKEM Security Model:")
	t.Log("- Tree stores only PUBLIC keys (safe to share)")
	t.Log("- Clients keep their PRIVATE keys locally")
	t.Log("- Intermediate public keys derived from children")
	t.Log("- Root public key = group's shared public key")
	
	t.Log("✓ TreeKEM public key derivation test completed successfully")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
