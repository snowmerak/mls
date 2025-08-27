package tree

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
	
	if !head.IsLeaf() {
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
	if head.IsLeaf() {
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
	
	i := 0
	for email, pubKey := range clientPublicKeys {
		err := tree.Insert(email, pubKey)
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", email, err)
		}
		
		// In TreeKEM, intermediate node public keys are set by clients after DH computation
		// The group public key (root) will be empty until clients compute and set it
		if i == 0 {
			// First user - tree head is a leaf, has public key
			groupPubKey := tree.GetGroupPublicKey()
			if len(groupPubKey) > 0 {
				t.Logf("  Group public key after adding %s: %x...", email, groupPubKey[:8])
			}
		} else {
			// Multiple users - root becomes intermediate node, needs client-side key computation
			t.Logf("  %s added, intermediate nodes need client-side key computation", email)
		}
		i++
	}
	
	// Check TreeKEM properties
	leaves := tree.GetLeaves()
	t.Logf("Found %d leaf nodes (actual users)", len(leaves))
	
	for _, leaf := range leaves {
		if leaf.nodeType != "leaf" {
			t.Errorf("Leaf node %s should have nodeType 'leaf'", leaf.name)
		}
		if len(leaf.publicKey) == 0 {
			t.Errorf("Leaf node %s should have a public key", leaf.name)
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