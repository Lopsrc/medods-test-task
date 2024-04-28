package mngstorage

import (
	"context"
	"errors"

	"test-task/internal/models"
	"test-task/internal/storage"
	"test-task/pkg/client/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	db mongo.Collection
}

func New(db mongo.Database) *Repository {
	return &Repository{
        db: *db.Collection(storage.UserCollection),
    }
}

func (r *Repository) Insert(ctx context.Context, GUID string) error {
	_, err := r.db.InsertOne(ctx, models.User{
		GUID: GUID,
	})
	if mongodb.IsDuplicate(err) {
		return storage.ErrAlreadyExists
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, user models.User) error {
	_, err := r.db.UpdateOne(ctx, 
		bson.M{"guid": user.GUID},
		bson.M{"$set": bson.M{"refresh_token_hash": user.RefreshTokenHash, "expiresAt":  user.ExpiresAt}},
    )
	return  err
}


func (r *Repository) Get(ctx context.Context, GUID string) ([]byte, error) {
	var user models.User

	err := r.db.FindOne(ctx, bson.M{"guid": GUID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, storage.ErrNotFound
		}
	}
	return user.RefreshTokenHash, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]models.User, error) {
    var users []models.User	
	cursor, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &users); err!= nil {
        return nil, err
    }
    return users, nil
}

