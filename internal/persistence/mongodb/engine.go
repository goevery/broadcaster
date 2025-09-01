package mongodb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/persistence"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Message struct {
	Id         bson.ObjectID `bson:"_id"`
	CreateTime time.Time
	ChannelId  string
	Payload    string
}

type PersistenceEngine struct {
	collection *mongo.Collection
}

func NewPersistenceEngine(client *mongo.Client) *PersistenceEngine {
	database := client.Database("broadcaster")
	collection := database.Collection("messages")

	return &PersistenceEngine{
		collection,
	}
}

func (e *PersistenceEngine) Setup(ctx context.Context) error {
	ttlIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "createTime", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(5 * 24 * 60 * 60),
	}

	channelIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "channelId", Value: 1},
			{Key: "_id", Value: -1},
		},
	}

	_, err := e.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{ttlIndexModel, channelIndexModel})

	return err
}

func (e *PersistenceEngine) Save(ctx context.Context, message persistence.SaveRequest) (broadcaster.Message, error) {
	createTime := time.Now()

	payloadJson, err := json.Marshal(message.Payload)
	if err != nil {
		return broadcaster.Message{}, err
	}

	result, err := e.collection.InsertOne(ctx, bson.D{
		{Key: "createTime", Value: createTime},
		{Key: "channelId", Value: message.ChannelId},
		{Key: "payload", Value: string(payloadJson)},
	})

	return broadcaster.Message{
		Id:         result.InsertedID.(bson.ObjectID).Hex(),
		CreateTime: createTime,
		ChannelId:  message.ChannelId,
		Payload:    message.Payload,
	}, err
}

func (e *PersistenceEngine) List(ctx context.Context, channelId string, lastSeenId string) ([]broadcaster.Message, error) {
	lastSeenObjectId, err := bson.ObjectIDFromHex(lastSeenId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id":       bson.M{"$gte": lastSeenObjectId},
		"channelId": channelId,
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(101)

	result, err := e.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var mongoMessages []Message
	err = result.All(ctx, &mongoMessages)
	if err != nil {
		return nil, err
	}

	messages := make([]broadcaster.Message, len(mongoMessages))
	for i, m := range mongoMessages {
		var payload any
		err := json.Unmarshal([]byte(m.Payload), &payload)
		if err != nil {
			return nil, err
		}

		messages[i] = broadcaster.Message{
			Id:         m.Id.Hex(),
			CreateTime: m.CreateTime,
			ChannelId:  m.ChannelId,
			Payload:    payload,
		}
	}

	return messages, nil
}
