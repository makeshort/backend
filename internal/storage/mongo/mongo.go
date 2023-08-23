package mongo

import (
	"backend/internal/config"
	"backend/internal/storage"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	CollectionUsers    = "users"
	CollectionUrls     = "urls"
	CollectionSessions = "sessions"
)

type Storage struct {
	Client          *mongo.Client
	config          *config.Config
	urls            *mongo.Collection
	users           *mongo.Collection
	refreshSessions *mongo.Collection
}

// New returns a new Storage instance
func New(cfg *config.Config) *Storage {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Db.ConnectionURI))
	if err != nil {
		panic(err)
	}

	var dbName string
	if cfg.Env == config.EnvProduction {
		dbName = cfg.Env
	} else {
		dbName = config.EnvDevelopment
	}

	db := client.Database(dbName)
	urls := db.Collection(CollectionUrls)
	_, err = urls.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "alias", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		panic(err)
	}

	users := db.Collection(CollectionUsers)
	_, err = users.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "username", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	refreshSessions := db.Collection(CollectionSessions)

	_, err = refreshSessions.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{"expires_at", 1}},
				Options: options.Index().SetExpireAfterSeconds(0),
			},
			{
				Keys:    bson.D{{"refresh_token", 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	return &Storage{Client: client, config: cfg, urls: urls, users: users, refreshSessions: refreshSessions}
}

// CreateURL creates a URL document in database
func (s *Storage) CreateURL(ctx context.Context, link string, alias string, userID primitive.ObjectID) (primitive.ObjectID, error) {
	doc, err := s.urls.InsertOne(ctx, storage.URL{
		Link:      link,
		Alias:     alias,
		UserID:    userID,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
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

// GetUrlByAlias returns a storage.URL object from database
// If url does not found, function will return a storage.ErrURLNotFound error
func (s *Storage) GetUrlByAlias(ctx context.Context, alias string) (storage.URL, error) {
	doc := s.urls.FindOne(ctx, bson.D{{"alias", alias}})
	var url storage.URL
	if err := doc.Decode(&url); err != nil {
		return storage.URL{}, storage.ErrURLNotFound
	}
	return url, nil
}

// IncrementRedirectsCounter increments redirects field of storage.URL document in database
func (s *Storage) IncrementRedirectsCounter(ctx context.Context, alias string) error {
	_, err := s.urls.UpdateOne(ctx,
		bson.D{{"alias", alias}},
		bson.D{{"$inc", bson.D{{"redirects", 1}}},
			{"$set", bson.D{{"updated_at", primitive.NewDateTimeFromTime(time.Now())}}}})
	if err != nil {
		return err
	}
	return nil
}

// DeleteURL deletes URL from database
func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	res, err := s.urls.DeleteOne(ctx, bson.D{{"alias", alias}})
	if res.DeletedCount == 0 {
		return storage.ErrURLNotFound
	}
	return err
}

// CreateUser creates a storage.User document in database
func (s *Storage) CreateUser(ctx context.Context, email string, username string, passwordHash string) (primitive.ObjectID, error) {
	doc, err := s.users.InsertOne(ctx, storage.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
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

// GetUserByCredentials returns a storage.User object from database
// If user does not found, function will return a storage.ErrUserNotFound error
func (s *Storage) GetUserByCredentials(ctx context.Context, email string, passwordHash string) (storage.User, error) {
	doc := s.users.FindOne(ctx, bson.D{{"email", email}, {"password_hash", passwordHash}})

	var user storage.User

	if err := doc.Decode(&user); err != nil {
		return storage.User{}, storage.ErrUserNotFound
	} // TODO: ErrNoDocuments check
	return user, nil
}

// GetUserURLs get and return all storage.URL documents in database, with given owner
func (s *Storage) GetUserURLs(ctx context.Context, userID primitive.ObjectID) ([]storage.URL, error) {
	cur, err := s.urls.Find(ctx, bson.D{{"user_id", userID}})
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

// DeleteUser deletes a user from database. If user does not found, function will return a storage.ErrUserNotFound error
func (s *Storage) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	res, err := s.users.DeleteOne(ctx, bson.D{{"_id", userID}})
	if res.DeletedCount == 0 {
		return storage.ErrUserNotFound
	}
	return err
}

// CreateRefreshSession creates a new refresh session with refresh token assigned to user
func (s *Storage) CreateRefreshSession(ctx context.Context, userID primitive.ObjectID, refreshToken string, ip string, userAgent string) (primitive.ObjectID, error) {
	doc, err := s.refreshSessions.InsertOne(ctx, storage.RefreshSession{
		UserID:       userID,
		RefreshToken: refreshToken,
		IP:           ip,
		UserAgent:    userAgent,
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		ExpiresAt:    primitive.NewDateTimeFromTime(time.Now().Add(s.config.Token.Refresh.TTL)),
	})
	if err != nil {
		return primitive.ObjectID{}, nil
	}

	return doc.InsertedID.(primitive.ObjectID), err
}

// DeleteRefreshSession deletes a refresh session from database
func (s *Storage) DeleteRefreshSession(ctx context.Context, refreshToken string) error {
	res, err := s.refreshSessions.DeleteOne(ctx, bson.D{{"refresh_token", refreshToken}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return storage.ErrRefreshSessionNotFound
	}
	return nil
}

// IsRefreshTokenValid checks is the refresh token has an active refresh session
func (s *Storage) IsRefreshTokenValid(ctx context.Context, refreshToken string) (isRefreshTokenValid bool, ownerID primitive.ObjectID) {
	var session storage.RefreshSession
	doc := s.refreshSessions.FindOne(ctx, bson.D{{"refresh_token", refreshToken}})
	if err := doc.Decode(&session); err != nil {
		return false, primitive.ObjectID{}
	}
	return true, session.UserID
}
