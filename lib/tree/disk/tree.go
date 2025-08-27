package disk

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/snowmerak/mls/lib/tree"
)

var _ tree.Element = &Element{}

type Element struct {
	name       string
	publicKey  []byte // TreeKEM public key (what goes in value field)
	leftCount  int
	rightCount int
	leftChild  *Element
	rightChild *Element
	filePath   string // disk storage path for this element

	// TreeKEM specific fields
	nodeType  string // "leaf" or "intermediate"
	leafIndex int    // for leaf nodes only
	nodeIndex int    // unique node number in the tree

	// Change tracking
	lastModified time.Time // 마지막 수정 시점
	lastChecked  time.Time // 마지막 확인 시점
}

// LeftChild implements tree.Element.
func (e *Element) LeftChild() tree.Element {
	if e.leftChild == nil {
		return nil
	}
	return e.leftChild
}

// LeftCount implements tree.Element.
func (e *Element) LeftCount() int {
	return e.leftCount
}

// Name implements tree.Element.
func (e *Element) Name() string {
	return e.name
}

// RightChild implements tree.Element.
func (e *Element) RightChild() tree.Element {
	if e.rightChild == nil {
		return nil
	}
	return e.rightChild
}

// RightCount implements tree.Element.
func (e *Element) RightCount() int {
	return e.rightCount
}

// SetLeftChild implements tree.Element.
func (e *Element) SetLeftChild(child tree.Element) {
	if child == nil {
		e.leftChild = nil
		return
	}

	if diskChild, ok := child.(*Element); ok {
		e.leftChild = diskChild
	} else {
		// Convert interface to disk element if needed
		e.leftChild = &Element{
			name:      child.Name(),
			publicKey: child.Value(),
		}
	}
}

// SetLeftCount implements tree.Element.
func (e *Element) SetLeftCount(count int) {
	e.leftCount = count
}

// SetRightChild implements tree.Element.
func (e *Element) SetRightChild(child tree.Element) {
	if child == nil {
		e.rightChild = nil
		return
	}

	if diskChild, ok := child.(*Element); ok {
		e.rightChild = diskChild
	} else {
		// Convert interface to disk element if needed
		e.rightChild = &Element{
			name:      child.Name(),
			publicKey: child.Value(),
		}
	}
}

// SetRightCount implements tree.Element.
func (e *Element) SetRightCount(count int) {
	e.rightCount = count
}

// Value implements tree.Element.
func (e *Element) Value() []byte {
	return e.publicKey
}

// NodeIndex returns the unique node number
func (e *Element) NodeIndex() int {
	return e.nodeIndex
}

// SetNodeIndex sets the unique node number
func (e *Element) SetNodeIndex(index int) {
	e.nodeIndex = index
}

// ParentIndex calculates parent node index
// TreeKEM convention: parent(n) = (n-1)/2 for n > 0
func (e *Element) ParentIndex() int {
	if e.nodeIndex == 0 {
		return -1 // root has no parent
	}
	return (e.nodeIndex - 1) / 2
}

// LeftChildIndex calculates left child index
// TreeKEM convention: left_child(n) = 2*n + 1
func (e *Element) LeftChildIndex() int {
	return 2*e.nodeIndex + 1
}

// RightChildIndex calculates right child index
// TreeKEM convention: right_child(n) = 2*n + 2
func (e *Element) RightChildIndex() int {
	return 2*e.nodeIndex + 2
}

// SiblingIndex calculates sibling node index
func (e *Element) SiblingIndex() int {
	if e.nodeIndex == 0 {
		return -1 // root has no sibling
	}
	if e.nodeIndex%2 == 1 {
		// left child, sibling is right child
		return e.nodeIndex + 1
	} else {
		// right child, sibling is left child
		return e.nodeIndex - 1
	}
}

// IsLeftChild checks if this node is a left child
func (e *Element) IsLeftChild() bool {
	return e.nodeIndex > 0 && e.nodeIndex%2 == 1
}

// IsRightChild checks if this node is a right child
func (e *Element) IsRightChild() bool {
	return e.nodeIndex > 0 && e.nodeIndex%2 == 0
}

// MarkAsModified updates the lastModified timestamp to current time
func (e *Element) MarkAsModified() {
	e.lastModified = time.Now()
}

// MarkAsChecked updates the lastChecked timestamp to current time
func (e *Element) MarkAsChecked() {
	e.lastChecked = time.Now()
}

// WasModifiedSince checks if the node was modified after the given time
func (e *Element) WasModifiedSince(since time.Time) bool {
	return e.lastModified.After(since)
}

// NeedsUpdate checks if the node needs to be updated (modified after last check)
func (e *Element) NeedsUpdate() bool {
	return e.lastModified.After(e.lastChecked)
}

// LastModified returns the last modification time
func (e *Element) LastModified() time.Time {
	return e.lastModified
}

// LastChecked returns the last check time
func (e *Element) LastChecked() time.Time {
	return e.lastChecked
}

// NewTree creates a new disk-based tree with the given root path.
func NewTree(rootPath string) (*Tree, error) {
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	return &Tree{
		rootPath: rootPath,
	}, nil
}

// LoadTree loads an existing tree from disk
func LoadTree(rootPath string, headName string) (*Tree, error) {
	tree := &Tree{
		rootPath: rootPath,
	}

	if headName != "" {
		headPath := tree.generateFilePath(headName)
		if _, err := os.Stat(headPath); err == nil {
			head, err := loadFromDisk(headPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load head element: %w", err)
			}
			tree.head = head
		}
	}

	return tree, nil
}

// elementData represents the serializable data for an element
type elementData struct {
	Name         string    `json:"name"`
	PublicKey    []byte    `json:"public_key"`
	LeftCount    int       `json:"left_count"`
	RightCount   int       `json:"right_count"`
	LeftChild    string    `json:"left_child,omitempty"`    // file path to left child
	RightChild   string    `json:"right_child,omitempty"`   // file path to right child
	NodeType     string    `json:"node_type"`               // "leaf" or "intermediate"
	LeafIndex    int       `json:"leaf_index,omitempty"`    // for leaf nodes only
	LastModified time.Time `json:"last_modified,omitempty"` // 마지막 수정 시점
	LastChecked  time.Time `json:"last_checked,omitempty"`  // 마지막 확인 시점
}

// SetValue updates the node's public key value
func (e *Element) SetValue(value []byte) {
	e.publicKey = value
}

// SaveToDisk is a public wrapper for saveToDisk
func (e *Element) SaveToDisk() error {
	return e.saveToDisk()
}

// saveToDisk saves the element to disk
func (e *Element) saveToDisk() error {
	if e.filePath == "" {
		return fmt.Errorf("element has no file path")
	}

	data := elementData{
		Name:         e.name,
		PublicKey:    e.publicKey,
		LeftCount:    e.leftCount,
		RightCount:   e.rightCount,
		NodeType:     e.nodeType,
		LeafIndex:    e.leafIndex,
		LastModified: e.lastModified,
		LastChecked:  e.lastChecked,
	}

	if e.leftChild != nil {
		data.LeftChild = e.leftChild.filePath
	}
	if e.rightChild != nil {
		data.RightChild = e.rightChild.filePath
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal element data: %w", err)
	}

	if err := os.WriteFile(e.filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write element to disk: %w", err)
	}

	return nil
}

// loadFromDisk loads an element from disk
func loadFromDisk(filePath string) (*Element, error) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read element from disk: %w", err)
	}

	var data elementData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal element data: %w", err)
	}

	element := &Element{
		name:         data.Name,
		publicKey:    data.PublicKey,
		leftCount:    data.LeftCount,
		rightCount:   data.RightCount,
		filePath:     filePath,
		nodeType:     data.NodeType,
		leafIndex:    data.LeafIndex,
		lastModified: data.LastModified,
		lastChecked:  data.LastChecked,
	}

	// Load children if they exist
	if data.LeftChild != "" {
		if leftChild, err := loadFromDisk(data.LeftChild); err == nil {
			element.leftChild = leftChild
		}
	}
	if data.RightChild != "" {
		if rightChild, err := loadFromDisk(data.RightChild); err == nil {
			element.rightChild = rightChild
		}
	}

	return element, nil
}

// generateFilePath generates a unique file path for an element
func (t *Tree) generateFilePath(name string) string {
	return filepath.Join(t.rootPath, fmt.Sprintf("%s.json", name))
}

var _ tree.Tree = &Tree{}

type Tree struct {
	rootPath      string   // base directory for storing tree data
	head          *Element // root element of the tree
	nextNodeIndex int      // counter for assigning unique node numbers
}

// Delete implements tree.Tree.
func (t *Tree) Delete(name string) error {
	if t.head == nil {
		return fmt.Errorf("tree is empty")
	}

	// Simple deletion: find the node and remove it, then compact the tree
	var deleteNode func(*Element, string) (*Element, bool, error)
	deleteNode = func(node *Element, targetName string) (*Element, bool, error) {
		if node == nil {
			return nil, false, nil
		}

		if node.name == targetName {
			// Found the node to delete - remove file
			if node.filePath != "" {
				os.Remove(node.filePath)
			}

			// Simple replacement strategy
			if node.leftChild == nil && node.rightChild == nil {
				return nil, true, nil
			}
			if node.leftChild == nil {
				return node.rightChild, true, nil
			}
			if node.rightChild == nil {
				return node.leftChild, true, nil
			}

			// Both children exist - promote left child and attach right as rightmost
			left := node.leftChild

			// Find rightmost position in left subtree to attach right subtree
			current := left
			for current.rightChild != nil {
				current = current.rightChild
			}
			current.rightChild = node.rightChild
			current.rightCount = node.rightChild.leftCount + node.rightChild.rightCount + 1
			current.saveToDisk()

			// Update counts
			left.rightCount = left.rightCount + current.rightCount
			left.saveToDisk()

			return left, true, nil
		}

		// Search in children
		var found bool
		var err error

		if node.leftChild != nil {
			node.leftChild, found, err = deleteNode(node.leftChild, targetName)
			if found {
				node.leftCount--
				node.saveToDisk()
				return node, true, err
			}
		}

		if node.rightChild != nil {
			node.rightChild, found, err = deleteNode(node.rightChild, targetName)
			if found {
				node.rightCount--
				node.saveToDisk()
				return node, true, err
			}
		}

		return node, false, nil
	}

	newHead, found, err := deleteNode(t.head, name)
	if !found {
		return fmt.Errorf("element not found: %s", name)
	}
	t.head = newHead

	// Reassign node indices and rename intermediate nodes after deletion
	// to maintain TreeKEM consistency
	t.renameIntermediateNodes()
	t.reassignNodeIndices()

	return err
}

// Find implements tree.Tree.
func (t *Tree) Find(name string) (tree.Element, bool) {
	// Breadth-first search since we're not using BST ordering
	if t.head == nil {
		return nil, false
	}

	// Use iterative approach to avoid stack overflow
	queue := []*Element{t.head}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.name == name {
			return current, true
		}

		if current.leftChild != nil {
			queue = append(queue, current.leftChild)
		}
		if current.rightChild != nil {
			queue = append(queue, current.rightChild)
		}
	}

	return nil, false
}

// Head implements tree.Tree.
func (t *Tree) Head() tree.Element {
	return t.head
}

// Insert implements tree.Tree.
// In TreeKEM, value is the user's public key
// This function only manages tree structure - actual key derivation happens client-side
func (t *Tree) Insert(name string, value []byte) error {
	newElement := &Element{
		name:         name,
		publicKey:    value, // This is the user's public key
		filePath:     t.generateFilePath(name),
		nodeType:     "leaf",
		leafIndex:    t.getNextLeafIndex(),
		nodeIndex:    t.nextNodeIndex, // assign unique node number
		lastModified: time.Now(),      // mark as modified when created
		lastChecked:  time.Time{},     // not checked yet
	}
	t.nextNodeIndex++ // increment for next node

	// Save new element to disk
	if err := newElement.saveToDisk(); err != nil {
		return fmt.Errorf("failed to save new element to disk: %w", err)
	}

	if t.head == nil {
		t.head = newElement
		t.head.SetNodeIndex(0) // root is always node 0
		t.nextNodeIndex = 1    // next node will be 1
		return nil
	}

	// TreeKEM insertion: only add to leaf positions
	// This approach creates a new intermediate parent when adding to a leaf
	var insertToLeaf func(**Element, *Element) error
	insertToLeaf = func(nodePtr **Element, newNode *Element) error {
		current := *nodePtr

		// Check if current node is a leaf (no children)
		if current.leftChild == nil && current.rightChild == nil {
			// This is a leaf - we need to split it
			// Create an intermediate node placeholder
			// In real TreeKEM, the public key would be provided by clients after DH computation
			intermediateNode := &Element{
				name:         fmt.Sprintf("intermediate_%s_%s", current.name, newNode.name),
				publicKey:    []byte{}, // Will be set by client-side key derivation
				filePath:     t.generateFilePath(fmt.Sprintf("intermediate_%s_%s", current.name, newNode.name)),
				leftChild:    current,
				rightChild:   newNode,
				leftCount:    1,
				rightCount:   1,
				nodeType:     "intermediate",
				nodeIndex:    t.nextNodeIndex, // assign unique node number
				lastModified: time.Now(),      // mark as modified when created
				lastChecked:  time.Time{},     // not checked yet
			}
			t.nextNodeIndex++ // increment for next node

			// Save intermediate node
			if err := intermediateNode.saveToDisk(); err != nil {
				return fmt.Errorf("failed to save intermediate node: %w", err)
			}

			// Replace current node's position with intermediate node
			*nodePtr = intermediateNode
			return nil
		}

		// Not a leaf - find the subtree with fewer leaves
		leftLeafCount := countLeaves(current.leftChild)
		rightLeafCount := countLeaves(current.rightChild)

		if leftLeafCount <= rightLeafCount {
			// Insert to left subtree
			if current.leftChild == nil {
				current.leftChild = newNode
				current.leftCount = 1
			} else {
				if err := insertToLeaf(&current.leftChild, newNode); err != nil {
					return err
				}
				current.leftCount++
			}
		} else {
			// Insert to right subtree
			if current.rightChild == nil {
				current.rightChild = newNode
				current.rightCount = 1
			} else {
				if err := insertToLeaf(&current.rightChild, newNode); err != nil {
					return err
				}
				current.rightCount++
			}
		}

		// In real TreeKEM, intermediate keys are set by clients, not automatically derived
		// We skip automatic key derivation here

		// Save updated current node
		return current.saveToDisk()
	}

	// Perform insertion
	if err := insertToLeaf(&t.head, newElement); err != nil {
		return err
	}

	// Reassign node indices to maintain TreeKEM ordering
	t.reassignNodeIndices()

	// In real TreeKEM, keys are set by clients after DH computation
	return nil
}

// Helper function to count leaf nodes in a subtree
func countLeaves(node *Element) int {
	if node == nil {
		return 0
	}

	// If it's a leaf node
	if node.leftChild == nil && node.rightChild == nil {
		return 1
	}

	// Count leaves in both subtrees
	return countLeaves(node.leftChild) + countLeaves(node.rightChild)
}

// TreeKEM public key derivation functions

// DerivePublicKey derives a public key for intermediate nodes
// In real TreeKEM, this would use proper cryptographic operations
// For now, we use a simple hash-based approach
// getNextLeafIndex returns the next available leaf index
func (t *Tree) getNextLeafIndex() int {
	if t.head == nil {
		return 0
	}

	leaves := t.GetLeaves()
	maxIndex := -1
	for _, leaf := range leaves {
		if element, ok := leaf.(*Element); ok {
			if element.leafIndex > maxIndex {
				maxIndex = element.leafIndex
			}
		}
	}
	return maxIndex + 1
}

// reassignNodeIndices assigns proper TreeKEM node indices to all nodes
// TreeKEM uses level-order (breadth-first) numbering: root=0, level1=[1,2], level2=[3,4,5,6], etc.
func (t *Tree) reassignNodeIndices() {
	if t.head == nil {
		return
	}

	// Use breadth-first traversal to assign indices
	queue := []*Element{t.head}
	index := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		current.SetNodeIndex(index)
		index++

		if current.leftChild != nil {
			queue = append(queue, current.leftChild)
		}
		if current.rightChild != nil {
			queue = append(queue, current.rightChild)
		}
	}

	t.nextNodeIndex = index
}

// renameIntermediateNodes updates intermediate node names after deletion
// to reflect the current leaf nodes in each subtree
func (t *Tree) renameIntermediateNodes() {
	if t.head == nil {
		return
	}

	var updateNames func(*Element)
	updateNames = func(node *Element) {
		if node == nil {
			return
		}

		// Recursively update children first
		updateNames(node.leftChild)
		updateNames(node.rightChild)

		// If this is an intermediate node, update its name
		if node.nodeType == "intermediate" {
			var leftLeafNames []string
			var rightLeafNames []string

			// Collect leaf names from left subtree
			if node.leftChild != nil {
				leftLeafNames = collectLeafNames(node.leftChild)
			}

			// Collect leaf names from right subtree
			if node.rightChild != nil {
				rightLeafNames = collectLeafNames(node.rightChild)
			}

			// Generate new name based on current leaves
			if len(leftLeafNames) > 0 && len(rightLeafNames) > 0 {
				oldFilePath := node.filePath
				newName := fmt.Sprintf("intermediate_%s_%s", leftLeafNames[0], rightLeafNames[0])
				node.name = newName
				node.filePath = t.generateFilePath(newName)

				// Remove old file and save with new name
				if oldFilePath != "" {
					os.Remove(oldFilePath)
				}
				node.saveToDisk()
			}
		}
	}

	updateNames(t.head)
}

// collectLeafNames collects all leaf node names in a subtree
func collectLeafNames(node *Element) []string {
	if node == nil {
		return nil
	}

	if node.nodeType == "leaf" {
		return []string{node.name}
	}

	var names []string
	names = append(names, collectLeafNames(node.leftChild)...)
	names = append(names, collectLeafNames(node.rightChild)...)
	return names
}

// GetNodeByIndex finds a node by its index number
func (t *Tree) GetNodeByIndex(targetIndex int) tree.Element {
	if t.head == nil {
		return nil
	}

	// Use breadth-first search to find the node
	queue := []*Element{t.head}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.NodeIndex() == targetIndex {
			return current
		}

		if current.leftChild != nil {
			queue = append(queue, current.leftChild)
		}
		if current.rightChild != nil {
			queue = append(queue, current.rightChild)
		}
	}

	return nil
}

func DerivePublicKey(leftPubKey, rightPubKey []byte) []byte {
	if len(leftPubKey) == 0 && len(rightPubKey) == 0 {
		return []byte{}
	}

	// Simple public key derivation using hash
	// In real TreeKEM, this would be more complex cryptographic operations
	hasher := sha256.New()

	// Add domain separation
	hasher.Write([]byte("TreeKEM-intermediate-pubkey"))

	// Add length prefixes to prevent collision attacks
	leftLen := make([]byte, 4)
	rightLen := make([]byte, 4)
	binary.BigEndian.PutUint32(leftLen, uint32(len(leftPubKey)))
	binary.BigEndian.PutUint32(rightLen, uint32(len(rightPubKey)))

	hasher.Write(leftLen)
	hasher.Write(leftPubKey)
	hasher.Write(rightLen)
	hasher.Write(rightPubKey)

	return hasher.Sum(nil)
}

// UpdateIntermediateKeys updates all intermediate node keys based on their children
// This should be called after any tree modification
func (t *Tree) UpdateIntermediateKeys() error {
	if t.head == nil {
		return nil
	}

	var updateKeys func(*Element) error
	updateKeys = func(node *Element) error {
		if node == nil {
			return nil
		}

		// Update children first (bottom-up)
		if node.leftChild != nil {
			if err := updateKeys(node.leftChild); err != nil {
				return err
			}
		}
		if node.rightChild != nil {
			if err := updateKeys(node.rightChild); err != nil {
				return err
			}
		}

		// If this is not a leaf, derive new public key from children
		if node.leftChild != nil || node.rightChild != nil {
			var leftPubKey, rightPubKey []byte

			if node.leftChild != nil {
				leftPubKey = node.leftChild.publicKey
			}
			if node.rightChild != nil {
				rightPubKey = node.rightChild.publicKey
			}

			// Derive new public key for this intermediate node
			node.publicKey = DerivePublicKey(leftPubKey, rightPubKey)

			// Save updated node
			if err := node.saveToDisk(); err != nil {
				return fmt.Errorf("failed to save updated intermediate node: %w", err)
			}
		}

		return nil
	}

	return updateKeys(t.head)
}

// GetGroupPublicKey returns the root public key of the tree (group public key in TreeKEM)
func (t *Tree) GetGroupPublicKey() []byte {
	if t.head == nil {
		return nil
	}
	return t.head.publicKey
}

// IsLeaf checks if a node is a leaf node (represents an actual user)
func (e *Element) IsLeaf() bool {
	return e.leftChild == nil && e.rightChild == nil
}

// GetLeaves returns all leaf nodes (actual users) in the tree
func (t *Tree) GetLeaves() []tree.Element {
	if t.head == nil {
		return nil
	}

	var leaves []tree.Element
	var collectLeaves func(*Element)
	collectLeaves = func(node *Element) {
		if node == nil {
			return
		}

		if node.IsLeaf() {
			leaves = append(leaves, node)
		} else {
			collectLeaves(node.leftChild)
			collectLeaves(node.rightChild)
		}
	}

	collectLeaves(t.head)
	return leaves
}

// GetPath returns the path from a leaf node to the root
// This is important for TreeKEM key derivation
func (t *Tree) GetPath(leafName string) ([]tree.Element, error) {
	if t.head == nil {
		return nil, fmt.Errorf("tree is empty")
	}

	var path []tree.Element
	var findPath func(*Element, string) bool
	findPath = func(node *Element, targetName string) bool {
		if node == nil {
			return false
		}

		// Add current node to path
		path = append(path, node)

		if node.name == targetName {
			return true
		}

		// Search in children
		if findPath(node.leftChild, targetName) || findPath(node.rightChild, targetName) {
			return true
		}

		// Remove from path if not found in this subtree
		path = path[:len(path)-1]
		return false
	}

	if findPath(t.head, leafName) {
		return path, nil
	}

	return nil, fmt.Errorf("leaf node not found: %s", leafName)
}

// SetIntermediateNodeKey allows clients to set the public key for an intermediate node
// after they have computed it using Diffie-Hellman key exchange
func (t *Tree) SetIntermediateNodeKey(nodeName string, publicKey []byte) error {
	node, found := t.Find(nodeName)
	if !found {
		return fmt.Errorf("node not found: %s", nodeName)
	}

	element, ok := node.(*Element)
	if !ok {
		return fmt.Errorf("invalid node type")
	}

	if element.nodeType != "intermediate" {
		return fmt.Errorf("can only set keys for intermediate nodes")
	}

	element.publicKey = publicKey
	element.MarkAsModified() // mark as modified when key is updated
	return element.saveToDisk()
}

// GetTreeStructure returns the current tree structure for client-side key computation
func (t *Tree) GetTreeStructure() map[string]*tree.NodeInfo {
	structure := make(map[string]*tree.NodeInfo)

	var traverse func(*Element)
	traverse = func(node *Element) {
		if node == nil {
			return
		}

		info := &tree.NodeInfo{
			Name:        node.name,
			PublicKey:   node.publicKey,
			NodeType:    node.nodeType,
			LeafIndex:   node.leafIndex,
			NodeIndex:   node.nodeIndex,
			ParentIndex: node.ParentIndex(),
		}

		if node.leftChild != nil {
			info.LeftChild = node.leftChild.name
		}
		if node.rightChild != nil {
			info.RightChild = node.rightChild.name
		}

		structure[node.name] = info

		traverse(node.leftChild)
		traverse(node.rightChild)
	}

	traverse(t.head)
	return structure
}

// GetModifiedNodes returns all nodes that have been modified since the given time
func (t *Tree) GetModifiedNodes(since time.Time) []tree.Element {
	if t.head == nil {
		return nil
	}

	var modifiedNodes []tree.Element
	var traverse func(*Element)
	traverse = func(node *Element) {
		if node == nil {
			return
		}

		if node.WasModifiedSince(since) {
			modifiedNodes = append(modifiedNodes, node)
		}

		traverse(node.leftChild)
		traverse(node.rightChild)
	}

	traverse(t.head)
	return modifiedNodes
}

// GetNodesNeedingUpdate returns all nodes that need updates (modified after last check)
func (t *Tree) GetNodesNeedingUpdate() []tree.Element {
	if t.head == nil {
		return nil
	}

	var needUpdateNodes []tree.Element
	var traverse func(*Element)
	traverse = func(node *Element) {
		if node == nil {
			return
		}

		if node.NeedsUpdate() {
			needUpdateNodes = append(needUpdateNodes, node)
		}

		traverse(node.leftChild)
		traverse(node.rightChild)
	}

	traverse(t.head)
	return needUpdateNodes
}

// MarkAllAsChecked marks all nodes in the tree as checked (updates lastChecked to now)
func (t *Tree) MarkAllAsChecked() {
	if t.head == nil {
		return
	}

	var traverse func(*Element)
	traverse = func(node *Element) {
		if node == nil {
			return
		}

		node.MarkAsChecked()
		node.saveToDisk() // persist the updated timestamp

		traverse(node.leftChild)
		traverse(node.rightChild)
	}

	traverse(t.head)
}

// GetNodeChangesSince returns a summary of nodes changed since the given time
func (t *Tree) GetNodeChangesSince(since time.Time) map[string]time.Time {
	changes := make(map[string]time.Time)

	if t.head == nil {
		return changes
	}

	var traverse func(*Element)
	traverse = func(node *Element) {
		if node == nil {
			return
		}

		if node.WasModifiedSince(since) {
			changes[node.name] = node.lastModified
		}

		traverse(node.leftChild)
		traverse(node.rightChild)
	}

	traverse(t.head)
	return changes
}
