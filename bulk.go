package goose

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Distinct returns distinct values for a field matching the filter.
// Equivalent to: await Model.distinct('status', filter)
func (m *Model[T]) Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	return m.Col().Distinct(ctx, fieldName, filter)
}

// BulkWrite executes multiple write operations in a single call.
// Equivalent to: await Model.bulkWrite([...operations])
//
// Usage:
//
//	results, err := MediaModel.BulkWrite(ctx, []mongo.WriteModel{
//	    mongo.NewInsertOneModel().SetDocument(bson.M{"type": "video"}),
//	    mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": id}).SetUpdate(bson.M{"$set": bson.M{"status": "done"}}),
//	    mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": oldID}),
//	})
func (m *Model[T]) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return m.Col().BulkWrite(ctx, models, opts...)
}

// EstimatedDocumentCount returns an estimate of the total number of documents.
// Faster than CountDocuments for large collections.
func (m *Model[T]) EstimatedDocumentCount(ctx context.Context) (int64, error) {
	return m.Col().EstimatedDocumentCount(ctx)
}
