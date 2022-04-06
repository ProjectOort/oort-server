package search

import "github.com/ProjectOort/oort-server/biz/search"

type Item struct {
	Type     int      `json:"type"`
	TargetID string   `json:"targetID"`
	Title    string   `json:"title"`
	Content  []string `json:"content"`
}

func MakeItemPresenter(item *search.Item) *Item {
	return &Item{
		Type:     item.Type,
		TargetID: item.TargetID,
		Title:    item.Title,
		Content:  item.Content,
	}
}
