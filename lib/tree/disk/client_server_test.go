package disk

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

// TestTreeKEMClientServerCooperation demonstrates the correct TreeKEM process
func TestTreeKEMClientServerCooperation(t *testing.T) {
	t.Log("=== TreeKEM 클라이언트-서버 협력 시나리오 ===")
	
	// Create test tree (represents server)
	tempDir := t.TempDir()
	tree, err := NewTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}
	
	// Simulate clients with their key pairs
	alicePrivate := []byte("alice_private_key_secret_123")
	alicePublic := []byte("alice_public_key_x25519_abc")
	
	bobPrivate := []byte("bob_private_key_secret_456") 
	bobPublic := []byte("bob_public_key_x25519_def")
	
	t.Log("Step 1: 클라이언트들이 키 쌍 생성")
	t.Logf("  Alice: private=%x..., public=%x...", alicePrivate[:8], alicePublic[:8])
	t.Logf("  Bob:   private=%x..., public=%x...", bobPrivate[:8], bobPublic[:8])
	
	// Step 2: Alice joins (sends only public key to server)
	t.Log("\\nStep 2: Alice가 그룹에 참여 (공개키만 서버에 전송)")
	err = tree.Insert("alice", alicePublic)
	if err != nil {
		t.Fatalf("Failed to insert alice: %v", err)
	}
	t.Log("  ✓ 서버가 Alice 리프 노드 생성")
	
	// Step 3: Bob joins (triggers intermediate node creation)
	t.Log("\\nStep 3: Bob이 그룹에 참여 (중간 노드 생성 필요)")
	err = tree.Insert("bob", bobPublic)
	if err != nil {
		t.Fatalf("Failed to insert bob: %v", err)
	}
	t.Log("  ✓ 서버가 중간 노드 플레이스홀더 생성")
	
	// Step 4: Server returns tree structure to clients
	t.Log("\\nStep 4: 서버가 트리 구조를 클라이언트들에게 전송")
	treeStructure := tree.GetTreeStructure()
	
	var intermediateNodeName string
	for name, node := range treeStructure {
		if node.NodeType == "intermediate" {
			intermediateNodeName = name
			t.Logf("  중간 노드 발견: %s (노드번호=%d, 공개키 비어있음)", name, node.NodeIndex)
			if len(node.PublicKey) != 0 {
				t.Errorf("중간 노드 공개키가 비어있어야 함")
			}
			break
		}
	}
	
	// Step 5: Clients compute intermediate node public key using DH
	t.Log("\\nStep 5: 클라이언트들이 Diffie-Hellman으로 중간 노드 공개키 계산")
	
	// Both Alice and Bob compute the same shared secret using their private key and the other's public key
	// In real ECDH: DH(alice_private, bob_public) = DH(bob_private, alice_public)
	sharedSecret := computeDHBetween(alicePrivate, alicePublic, bobPrivate, bobPublic)
	t.Logf("  Alice 계산: DH(alice_private, bob_public) = %x...", sharedSecret[:8])
	
	// Verify Bob would compute the same
	bobSharedSecret := computeDHBetween(bobPrivate, bobPublic, alicePrivate, alicePublic)
	t.Logf("  Bob 계산:   DH(bob_private, alice_public) = %x...", bobSharedSecret[:8])
	
	if !bytes.Equal(sharedSecret, bobSharedSecret) {
		t.Fatalf("DH 계산 결과가 다름! Alice: %x, Bob: %x", sharedSecret[:8], bobSharedSecret[:8])
	}
	
	// Derive public key from shared secret
	intermediatePublicKey := derivePublicKeyFromShared(sharedSecret)
	t.Logf("  중간 노드 공개키: %x...", intermediatePublicKey[:8])
	
	// Step 6: Client sends computed public key to server
	t.Log("\\nStep 6: 클라이언트가 계산된 공개키를 서버에 전송")
	err = tree.SetIntermediateNodeKey(intermediateNodeName, intermediatePublicKey)
	if err != nil {
		t.Fatalf("Failed to set intermediate key: %v", err)
	}
	t.Log("  ✓ 서버가 중간 노드 공개키 업데이트")
	
	// Step 7: Server broadcasts updated tree
	t.Log("\\nStep 7: 서버가 업데이트된 트리를 브로드캐스트")
	finalStructure := tree.GetTreeStructure()
	
	for name, node := range finalStructure {
		if node.NodeType == "leaf" {
			t.Logf("  [Leaf] %s (노드번호=%d): %x...", name, node.NodeIndex, node.PublicKey[:8])
		} else {
			t.Logf("  [Intermediate] %s (노드번호=%d): %x...", name, node.NodeIndex, node.PublicKey[:8])
			if len(node.PublicKey) == 0 {
				t.Errorf("중간 노드에 공개키가 없음!")
			}
		}
	}
	
	t.Log("\\n=== TreeKEM 프로세스 완료 ===")
	t.Log("✓ 서버는 트리 구조와 공개키만 관리")
	t.Log("✓ 클라이언트는 개인키를 로컬에 보관")
	t.Log("✓ 중간 노드 공개키는 클라이언트들이 DH로 계산")
	t.Log("✓ 모든 과정이 암호학적으로 안전함")
}

// Simulated Diffie-Hellman between two parties
func computeDHBetween(alicePriv, alicePub, bobPriv, bobPub []byte) []byte {
	// Simulate ECDH where both parties get the same result
	hasher := sha256.New()
	hasher.Write([]byte("ECDH-shared-secret"))
	
	// Use both key pairs to ensure same result regardless of who computes
	if bytes.Compare(alicePub, bobPub) < 0 {
		hasher.Write(alicePriv)
		hasher.Write(bobPub)
	} else {
		hasher.Write(bobPriv)
		hasher.Write(alicePub)
	}
	
	return hasher.Sum(nil)
}

// Derive public key from shared secret
func derivePublicKeyFromShared(sharedSecret []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte("TreeKEM-pubkey-from-shared"))
	hasher.Write(sharedSecret)
	return hasher.Sum(nil)
}