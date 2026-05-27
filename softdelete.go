package goose

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ── Soft Delete ──────────────────────────────────────────────
//
// Instead of permanently deleting documents, marks them as deleted
// with a deletedAt timestamp. Provides methods to query only
// non-deleted or only deleted documents.
//
// Usage:
//
//	type Post struct {
//	    goose.BaseModel     `bson:",inline"`
//	    goose.SoftDeleteModel `bson:",inline"`
//	    Title string          `bson:"title"`
//	}
//
//	// Soft delete
//	PostModel.SoftDelete(ctx, bson.M{"_id": id})
//
//	// Find only non-deleted
//	PostModel.FindActive(ctx, bson.M{"status": "draft"})
//
//	// Restore
//	PostModel.Restore(ctx, bson.M{"_id": id})

// SoftDeleteModel provides soft delete fields.
// Embed this in your model structs with `bson:",inline"`.
type SoftDeleteModel struct {
	DeletedAt *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}

// IsDeleted returns true if the document has been soft-deleted.
func (s *SoftDeleteModel) IsDeleted() bool {
	return s.DeletedAt != nil
}

// SoftDelete marks documents as deleted by setting deletedAt = now.
// Equivalent to: await Model.updateOne(filter, { $set: { deletedAt: new Date() } })
func (m *Model[T]) SoftDelete(ctx context.Context, filter interface{}) (*mongo.UpdateResult, error) {
	now := time.Now()
	return m.Col().UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{
			"deletedAt": now,
			"updatedAt": now,
		},
	})
}

// SoftDeleteByID soft-deletes a document by _id.
func (m *Model[T]) SoftDeleteByID(ctx context.Context, id string) (*mongo.UpdateResult, error) {
	return m.SoftDelete(ctx, bson.M{"_id": id})
}

// SoftDeleteMany marks multiple documents as deleted.
func (m *Model[T]) SoftDeleteMany(ctx context.Context, filter interface{}) (*mongo.UpdateResult, error) {
	now := time.Now()
	return m.Col().UpdateMany(ctx, filter, bson.M{
		"$set": bson.M{
			"deletedAt": now,
			"updatedAt": now,
		},
	})
}

// Restore removes the deletedAt field, un-deleting the document.
func (m *Model[T]) Restore(ctx context.Context, filter interface{}) (*mongo.UpdateResult, error) {
	return m.Col().UpdateOne(ctx, filter, bson.M{
		"$unset": bson.M{"deletedAt": ""},
		"$set":   bson.M{"updatedAt": time.Now()},
	})
}

// RestoreByID restores a soft-deleted document by _id.
func (m *Model[T]) RestoreByID(ctx context.Context, id string) (*mongo.UpdateResult, error) {
	return m.Restore(ctx, bson.M{"_id": id})
}

// RestoreMany restores multiple soft-deleted documents.
func (m *Model[T]) RestoreMany(ctx context.Context, filter interface{}) (*mongo.UpdateResult, error) {
	return m.Col().UpdateMany(ctx, filter, bson.M{
		"$unset": bson.M{"deletedAt": ""},
		"$set":   bson.M{"updatedAt": time.Now()},
	})
}

// ── Query helpers for soft delete ───────────────────────────

// activeFilter merges the user filter with { deletedAt: null } to exclude soft-deleted docs.
func activeFilter(filter interface{}) bson.M {
	if filter == nil {
		return bson.M{"deletedAt": nil}
	}
	if m, ok := filter.(bson.M); ok {
		m["deletedAt"] = nil
		return m
	}
	// Wrap non-bson.M filters
	return bson.M{"$and": []interface{}{filter, bson.M{"deletedAt": nil}}}
}

// deletedFilter merges the user filter with { deletedAt: { $ne: null } }.
func deletedFilter(filter interface{}) bson.M {
	if filter == nil {
		return bson.M{"deletedAt": bson.M{"$ne": nil}}
	}
	if m, ok := filter.(bson.M); ok {
		m["deletedAt"] = bson.M{"$ne": nil}
		return m
	}
	return bson.M{"$and": []interface{}{filter, bson.M{"deletedAt": bson.M{"$ne": nil}}}}
}

// FindActive finds documents that have NOT been soft-deleted.
// Automatically adds { deletedAt: null } to the filter.
func (m *Model[T]) FindActive(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*T, error) {
	return m.Find(ctx, activeFilter(filter), opts...)
}

// FindOneActive finds a single non-deleted document.
func (m *Model[T]) FindOneActive(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*T, error) {
	return m.FindOne(ctx, activeFilter(filter), opts...)
}

// FindDeleted finds only soft-deleted documents.
func (m *Model[T]) FindDeleted(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*T, error) {
	return m.Find(ctx, deletedFilter(filter), opts...)
}

// CountActive counts only non-deleted documents.
func (m *Model[T]) CountActive(ctx context.Context, filter interface{}) (int64, error) {
	return m.CountDocuments(ctx, activeFilter(filter))
}

// CountDeleted counts only soft-deleted documents.
func (m *Model[T]) CountDeleted(ctx context.Context, filter interface{}) (int64, error) {
	return m.CountDocuments(ctx, deletedFilter(filter))
}

// ExistsActive checks if a non-deleted document exists.
func (m *Model[T]) ExistsActive(ctx context.Context, filter interface{}) (bool, error) {
	return m.Exists(ctx, activeFilter(filter))
}

// QueryActive starts a chainable query that excludes soft-deleted documents.
func (m *Model[T]) QueryActive(filter interface{}) *Query[T] {
	return m.Query(activeFilter(filter))
}

// PermanentDelete hard-deletes (permanently removes) documents.
// Use this instead of DeleteOne when you truly want to remove data.
func (m *Model[T]) PermanentDelete(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return m.Col().DeleteOne(ctx, filter)
}

// PermanentDeleteMany hard-deletes multiple documents permanently.
func (m *Model[T]) PermanentDeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return m.Col().DeleteMany(ctx, filter)
}
