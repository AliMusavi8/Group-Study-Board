package db

import (
	"context"
	"errors"
	"time"

	"groupstudyboard/internal/config"
	"groupstudyboard/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	client    *mongo.Client
	db        *mongo.Database
	events    *mongo.Collection
	snapshots *mongo.Collection
	boards    *mongo.Collection
}

func NewStore(ctx context.Context, cfg config.Config) (*Store, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.DatabaseName)
	store := &Store{
		client:    client,
		db:        db,
		events:    db.Collection("board_events"),
		snapshots: db.Collection("board_snapshots"),
		boards:    db.Collection("boards"),
	}

	if err := store.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) ensureIndexes(ctx context.Context) error {
	_, err := s.events.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "roomId", Value: 1}, {Key: "createdAt", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "roomId", Value: 1}, {Key: "seq", Value: 1}},
		},
	})
	if err != nil {
		return err
	}

	_, err = s.snapshots.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "roomId", Value: 1}, {Key: "createdAt", Value: -1}},
	})
	if err != nil {
		return err
	}

	_, err = s.boards.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "roomId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (s *Store) EnsureRoom(ctx context.Context, roomID string) error {
	_, err := s.boards.UpdateOne(ctx,
		bson.M{"roomId": roomID},
		bson.M{"$setOnInsert": bson.M{"roomId": roomID, "createdAt": time.Now()}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (s *Store) SaveEvent(ctx context.Context, event models.Event) error {
	event.CreatedAt = time.Now()
	_, err := s.events.InsertOne(ctx, event)
	return err
}

func (s *Store) LoadSnapshot(ctx context.Context, roomID string) (*models.SnapshotPayload, time.Time, error) {
	var result struct {
		RoomID    string               `bson:"roomId"`
		Events    []models.Event       `bson:"events"`
		CreatedAt time.Time            `bson:"createdAt"`
	}

	err := s.snapshots.FindOne(ctx,
		bson.M{"roomId": roomID},
		options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	).Decode(&result)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, err
	}

	snapshot := &models.SnapshotPayload{
		Events:    result.Events,
		CreatedAt: result.CreatedAt.UnixMilli(),
	}
	return snapshot, result.CreatedAt, nil
}

func (s *Store) LoadEventsSince(ctx context.Context, roomID string, since time.Time) ([]models.Event, error) {
	filter := bson.M{"roomId": roomID}
	if !since.IsZero() {
		filter["createdAt"] = bson.M{"$gt": since}
	}

	cursor, err := s.events.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "seq", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.Event
	for cursor.Next(ctx) {
		var ev models.Event
		if err := cursor.Decode(&ev); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Store) CreateSnapshot(ctx context.Context, roomID string) error {
	cursor, err := s.events.Find(ctx, bson.M{"roomId": roomID}, options.Find().SetSort(bson.D{{Key: "seq", Value: 1}}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var events []models.Event
	for cursor.Next(ctx) {
		var ev models.Event
		if err := cursor.Decode(&ev); err != nil {
			return err
		}
		events = append(events, ev)
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	snapshot := bson.M{
		"roomId":    roomID,
		"events":    events,
		"createdAt": time.Now(),
	}
	_, err = s.snapshots.InsertOne(ctx, snapshot)
	return err
}

func (s *Store) LoadSnapshotAndEvents(ctx context.Context, roomID string) (*models.SnapshotPayload, []models.Event, error) {
	snapshot, snapshotTime, err := s.LoadSnapshot(ctx, roomID)
	if err != nil {
		return nil, nil, err
	}
	var since time.Time
	if snapshot != nil {
		since = snapshotTime
	}
	events, err := s.LoadEventsSince(ctx, roomID, since)
	if err != nil {
		return snapshot, nil, err
	}
	return snapshot, events, nil
}
