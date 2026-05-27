package goose

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DeleteOne deletes a single document.
// Runs pre("delete") filter hooks → delete → post("delete") filter hooks.
// Equivalent to: await MediaModel.deleteOne(filter)
func (m *Model[T]) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	if err := m.runFilterHooks(ctx, HookPre, EventDelete, filter); err != nil {
		return nil, err
	}
	result, err := m.Col().DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	_ = m.runFilterHooks(ctx, HookPost, EventDelete, filter)
	return result, nil
}

// DeleteByID deletes a document by _id.
func (m *Model[T]) DeleteByID(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	return m.DeleteOne(ctx, bson.M{"_id": id})
}

// DeleteMany deletes multiple documents.
// Runs pre("delete") filter hooks → delete → post("delete") filter hooks.
// Equivalent to: await MediaModel.deleteMany(filter)
func (m *Model[T]) DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	if err := m.runFilterHooks(ctx, HookPre, EventDelete, filter); err != nil {
		return nil, err
	}
	result, err := m.Col().DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	_ = m.runFilterHooks(ctx, HookPost, EventDelete, filter)
	return result, nil
}
