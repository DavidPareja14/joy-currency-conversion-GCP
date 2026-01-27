package main

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type PubSubNotifier struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewPubSubNotifier(projectID, topicID string) (*PubSubNotifier, error) {
	ctx := context.Background()
	
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	topic := client.Topic(topicID)
	
	return &PubSubNotifier{
		client: client,
		topic:  topic,
	}, nil
}

func (n *PubSubNotifier) SendNotification(ctx context.Context, notification EmailNotification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	result := n.topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("error publishing message: %w", err)
	}

	return nil
}

func (n *PubSubNotifier) Close() error {
	n.topic.Stop()
	return n.client.Close()
}