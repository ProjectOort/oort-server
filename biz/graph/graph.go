package graph

type Graph struct {
	Nodes []Node
	Links []Link
}

type Node struct {
	ID    string
	Hub   bool
	Title string
}

type Link struct {
	Source string
	Target string
}
