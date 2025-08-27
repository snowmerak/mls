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
}

type Tree interface {
	Head() Element
	Insert(name string, value []byte) error
	Find(name string) (Element, bool)
	Delete(name string) error
}
