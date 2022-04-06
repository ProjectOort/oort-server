package graph

import "github.com/ProjectOort/oort-server/biz/graph"

type Graph struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

type Node struct {
	ID    string `json:"id"`
	Hub   bool   `json:"hub"`
	Title string `json:"title"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func MakeGraphPresenter(gph *graph.Graph) *Graph {
	g := &Graph{
		Nodes: make([]Node, len(gph.Nodes)),
		Links: make([]Link, len(gph.Links)),
	}
	for i, node := range gph.Nodes {
		g.Nodes[i] = Node{
			ID:    node.ID,
			Hub:   node.Hub,
			Title: node.Title,
		}
	}
	for i, link := range gph.Links {
		g.Links[i] = Link{
			Source: link.Source,
			Target: link.Target,
		}
	}
	return g
}
