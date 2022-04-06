package asteroid

import (
	"time"

	"github.com/ProjectOort/oort-server/biz/asteroid"
)

type Asteroid struct {
	ID          string    `json:"id"`
	Hub         bool      `json:"hub"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

func MakeAsteroidPresenter(ast *asteroid.Asteroid) *Asteroid {
	return &Asteroid{
		ID:          ast.ID.Hex(),
		Hub:         ast.Hub,
		Title:       ast.Title,
		Content:     ast.Content,
		CreatedTime: ast.CreatedTime,
		UpdatedTime: ast.UpdatedTime,
	}
}

type Item struct {
	ID    string `json:"id"`
	Hub   bool   `json:"hub"`
	Title string `json:"title"`
}

func MakeItemPresenter(ast *asteroid.Asteroid) *Item {
	return &Item{
		ID:    ast.ID.Hex(),
		Hub:   ast.Hub,
		Title: ast.Title,
	}
}
