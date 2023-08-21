package mongo

import (
	"backend/internal/storage"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Mongo struct {
	Client   *mongo.Client
	urls     *mongo.Collection
	users    *mongo.Collection
	sessions *mongo.Collection
}

// New returns a new Mongo instance.
func New(mongoURI string, env string) *Mongo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	var dbName string
	if env == "prod" {
		dbName = env
	} else {
		dbName = "dev"
	}

	db := client.Database(dbName)
	urls := db.Collection("urls")
	_, err = urls.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "alias", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}

	users := db.Collection("users")
	_, err = users.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}

	sessions := db.Collection("sessions")

	return &Mongo{Client: client, urls: urls, users: users, sessions: sessions}
}

// CreateURL creates a URL document in database.
func (m *Mongo) CreateURL(ctx context.Context, link string, alias string, userID primitive.ObjectID) (primitive.ObjectID, error) {
	datetime := getPrimitiveDatetimeNow()
	doc, err := m.urls.InsertOne(ctx, storage.URL{
		Link:      link,
		Alias:     alias,
		UserID:    userID,
		CreatedAt: datetime,
		UpdatedAt: datetime,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return primitive.ObjectID{}, storage.ErrAliasAlreadyExists
		} else {
			return primitive.ObjectID{}, err
		}
	}

	return doc.InsertedID.(primitive.ObjectID), err
}

// GetURL returns a storage.URL object from database.
// If url does not found, function will return a storage.ErrURLNotFound error.
func (m *Mongo) GetURL(ctx context.Context, alias string) (storage.URL, error) {
	doc := m.urls.FindOne(ctx, bson.D{{"alias", alias}})
	var url storage.URL
	if err := doc.Decode(&url); err != nil {
		return storage.URL{}, storage.ErrURLNotFound
	}
	return url, nil
}

// IncrementUrlCounter increments redirects field of storage.URL document in database.
func (m *Mongo) IncrementUrlCounter(ctx context.Context, alias string) error {
	datetime := getPrimitiveDatetimeNow()
	_, err := m.urls.UpdateOne(ctx,
		bson.D{{"alias", alias}},
		bson.D{{"$inc", bson.D{{"redirects", 1}}},
			{"$set", bson.D{{"updated_at", datetime}}}})
	if err != nil {
		return err
	}
	return nil
}

// DeleteURL deletes URL from database.
func (m *Mongo) DeleteURL(ctx context.Context, alias string) error {
	res, err := m.urls.DeleteOne(ctx, bson.D{{"alias", alias}})
	if res.DeletedCount == 0 {
		return storage.ErrURLNotFound
	}
	return err
}

// CreateUser creates a storage.User document in database.
func (m *Mongo) CreateUser(ctx context.Context, email string, username string, passwordHash string) (primitive.ObjectID, error) {
	datetime := getPrimitiveDatetimeNow()
	doc, err := m.users.InsertOne(ctx, storage.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    datetime,
		UpdatedAt:    datetime,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return primitive.ObjectID{}, storage.ErrUserAlreadyExists
		} else {
			return primitive.ObjectID{}, err
		}
	}

	return doc.InsertedID.(primitive.ObjectID), err
}

// GetUser returns a storage.User object from database.
// If user does not found, function will return a storage.ErrUserNotFound error.
func (m *Mongo) GetUser(ctx context.Context, email string, passwordHash string) (storage.User, error) {
	doc := m.users.FindOne(ctx, bson.D{{"email", email}, {"password_hash", passwordHash}})

	var user storage.User

	if err := doc.Decode(&user); err != nil {
		return storage.User{}, storage.ErrUserNotFound
	}
	return user, nil
}

// GetUserURLs get and return all storage.URL documents in database, with given owner.
func (m *Mongo) GetUserURLs(ctx context.Context, userID primitive.ObjectID) ([]storage.URL, error) {
	cur, err := m.urls.Find(ctx, bson.D{{"user_id", userID}})
	if err != nil {
		return nil, err
	}
	var results []storage.URL
	for cur.Next(ctx) {
		var res storage.URL
		if err = cur.Decode(&res); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	if results == nil {
		return []storage.URL{}, nil
	}
	return results, nil
}

// DeleteUser deletes a user from database. If user does not found, function will return a storage.ErrUserNotFound error.
func (m *Mongo) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	res, err := m.users.DeleteOne(ctx, bson.D{{"_id", userID}})
	if res.DeletedCount == 0 {
		return storage.ErrUserNotFound
	}
	return err
}

func getPrimitiveDatetimeNow() primitive.DateTime {
	return primitive.NewDateTimeFromTime(time.Now())
}
