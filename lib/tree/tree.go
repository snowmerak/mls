package tree

type Element interface {
	Name() string
	Value() []byte
	LeftCount() int
	RightCount() int
	LeftChild() Element
	RightChild() Element
	SetLeftChild(Element)
	SetRightChild(Element)
	SetLeftCount(int)
	SetRightCount(int)
	SetValue([]byte)

	// TreeKEM node indexing methods
	NodeIndex() int
	SetNodeIndex(int)
	ParentIndex() int
	LeftChildIndex() int
	RightChildIndex() int
	SiblingIndex() int
	IsLeftChild() bool
	IsRightChild() bool
	MarkAsModified()
	MarkAsChecked()
}

type Tree interface {
	Head() Element
	Insert(name string, value []byte) error
	Find(name string) (Element, bool)
	Delete(name string) error

	// TreeKEM node management methods
	GetNodeByIndex(index int) Element
	GetTreeStructure() map[string]*NodeInfo
	SetIntermediateNodeKey(name string, publicKey []byte) error
}

// NodeInfo represents tree node information for TreeKEM coordination
type NodeInfo struct {
	Name        string `json:"name"`
	PublicKey   []byte `json:"public_key"`
	NodeType    string `json:"node_type"`
	LeafIndex   int    `json:"leaf_index"`
	NodeIndex   int    `json:"node_index"`
	ParentIndex int    `json:"parent_index"`
	LeftChild   string `json:"left_child,omitempty"`
	RightChild  string `json:"right_child,omitempty"`
}
