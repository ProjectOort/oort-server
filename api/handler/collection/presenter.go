package collection

import (
	"github.com/ProjectOort/oort-server/biz/collection"
	"time"
)

type Collection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Hub         bool      `json:"hub"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

func MakeCollectionPresenter(col *collection.Collection) *Collection {
	return &Collection{
		ID:          col.ID.Hex(),
		Name:        col.Name,
		Description: col.Description,
	}
}

func MakeItemPresenter(item *collection.Item) *Item {
	return &Item{
		ID:          item.ID.Hex(),
		Title:       item.Title,
		Hub:         item.Hub,
		CreatedTime: item.CreatedTime,
		UpdatedTime: item.UpdatedTime,
	}
}
