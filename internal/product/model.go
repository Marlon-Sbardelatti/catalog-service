package product

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Product struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID  bson.ObjectID `bson:"category_id" json:"category_id"`
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	Price       int64         `bson:"price" json:"price"`
	ImageKey    string        `json:"-" bson:"image_key"`
	ImageURL    string        `json:"image_url" bson:"-"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at"`
}
