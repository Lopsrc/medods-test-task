package mongodb

import (
	"context"
	"errors"
	"fmt"
	"test-task/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(ctx context.Context, sc config.StorageConfig, authDB string) (db *mongo.Database, err error) {
	var mongoDBURL string
	var isAuth bool
	if sc.Username == "" && sc.Password == "" {
		mongoDBURL = fmt.Sprintf("mongodb://%s:%s", sc.Host, sc.Port)
	} else {
		isAuth = true
		mongoDBURL = fmt.Sprintf("mongodb://%s:%s@%s:%s", sc.Username, sc.Password, sc.Host, sc.Port)
	}

	clientOptions := options.Client().ApplyURI(mongoDBURL)
	if isAuth {
		if authDB == "" {
			authDB = sc.Database
		}
		clientOptions.SetAuth(options.Credential{
			AuthSource: authDB,
			Username:   sc.Username,
			Password:   sc.Password,
		})
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongoDB due to error: %v", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongoDB due to error: %v", err)
	}

	return client.Database(sc.Database), nil
}

func IsDuplicate(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}

	return false
}