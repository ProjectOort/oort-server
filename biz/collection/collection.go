package collection

import (
	"github.com/ProjectOort/oort-server/biz/asteroid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Collection struct {
	ID          primitive.ObjectID   `bson:"_id"`
	State       bool                 `bson:"state"`
	CreatedTime time.Time            `bson:"created_time"`
	UpdatedTime time.Time            `bson:"updated_time"`
	Name        string               `bson:"name"`
	Description string               `bson:"description"`
	OwnerID     primitive.ObjectID   `bson:"owner_id"`
	Items       []primitive.ObjectID `bson:"items"`
}

type Item struct {
	asteroid.Asteroid
}
