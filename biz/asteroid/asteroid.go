package asteroid

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asteroid struct {
	ID          primitive.ObjectID `bson:"_id"`
	State       bool               `bson:"state"`
	CreatedTime time.Time          `bson:"created_time"`
	UpdatedTime time.Time          `bson:"updated_time"`

	AuthorID primitive.ObjectID `bson:"author_id"`
	Hub      bool               `bson:"hub"`
	Type     int                `bson:"type"`

	Title   string `bson:"title"`
	Content string `bson:"content"`
}
