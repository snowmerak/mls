package main

import (
	"log"

	"github.com/snowmerak/mls/lib/tree"
)

func main() {
	tree, err := tree.NewTree("./data")
	if err != nil {
		panic(err)
	}

	tree.Insert("user_1", []byte("User 1 key"))
	tree.Insert("user_2", []byte("User 2 key"))
	tree.Insert("user_3", []byte("User 3 key"))
	tree.Insert("user_4", []byte("User 4 key"))
	tree.Insert("user_5", []byte("User 5 key"))

	n, ok := tree.Find("user_5")
	if ok {
		n.SetValue([]byte("Updated User 5 key"))
	}

	needToUpdate := tree.GetNodesNeedingUpdate()
	for _, node := range needToUpdate {
		log.Printf("Node %s needs update: %+v", node.Name(), node)
	}
}
