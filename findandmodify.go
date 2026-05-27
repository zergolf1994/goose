package goose

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindOneAndUpdate atomically finds a document and updates it, returning the document.
// By default returns the document AFTER the update (unlike Mongoose which returns before).
// Equivalent to: await Model.findOneAndUpdate(filter, update, { new: true })
func (m *Model[T]) FindOneAndUpdate(ctx context.Context, filter interface{}, update bson.M, opts ...*options.FindOneAndUpdateOptions) (*T, error) {
	injectUpdatedAt(update)

	// Default: return document after update (like Mongoose { new: true })
	merged := options.FindOneAndUpdate().SetReturnDocument(options.After)
	for _, o := range opts {
		if o.ReturnDocument != nil {
			merged.SetReturnDocument(*o.ReturnDocument)
		}
		if o.Upsert != nil {
			merged.SetUpsert(*o.Upsert)
		}
		if o.Projection != nil {
			merged.SetProjection(o.Projection)
		}
		if o.Sort != nil {
			merged.SetSort(o.Sort)
		}
	}

	doc := new(T)
	err := m.Col().FindOneAndUpdate(ctx, filter, update, merged).Decode(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// FindByIDAndUpdate atomically finds a document by _id and updates it.
// Equivalent to: await Model.findByIdAndUpdate(id, update, { new: true })
func (m *Model[T]) FindByIDAndUpdate(ctx context.Context, id string, update bson.M, opts ...*options.FindOneAndUpdateOptions) (*T, error) {
	return m.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts...)
}

// FindOneAndDelete atomically finds a document and deletes it, returning the deleted document.
// Equivalent to: await Model.findOneAndDelete(filter)
func (m *Model[T]) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) (*T, error) {
	doc := new(T)
	err := m.Col().FindOneAndDelete(ctx, filter, opts...).Decode(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// FindByIDAndDelete atomically finds a document by _id and deletes it.
// Equivalent to: await Model.findByIdAndDelete(id)
func (m *Model[T]) FindByIDAndDelete(ctx context.Context, id string) (*T, error) {
	return m.FindOneAndDelete(ctx, bson.M{"_id": id})
}

// FindOneAndReplace atomically finds a document and replaces it entirely.
// Equivalent to: await Model.findOneAndReplace(filter, replacement)
func (m *Model[T]) FindOneAndReplace(ctx context.Context, filter interface{}, replacement *T, opts ...*options.FindOneAndReplaceOptions) (*T, error) {
	// Touch updatedAt if BaseModel is embedded
	if base := getBaseModel(replacement); base != nil {
		base.UpdatedAt = time.Now()
	}

	merged := options.FindOneAndReplace().SetReturnDocument(options.After)
	for _, o := range opts {
		if o.ReturnDocument != nil {
			merged.SetReturnDocument(*o.ReturnDocument)
		}
		if o.Upsert != nil {
			merged.SetUpsert(*o.Upsert)
		}
		if o.Projection != nil {
			merged.SetProjection(o.Projection)
		}
	}

	doc := new(T)
	err := m.Col().FindOneAndReplace(ctx, filter, replacement, merged).Decode(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Upsert performs an upsert: update if exists, insert if not.
// Equivalent to: await Model.updateOne(filter, update, { upsert: true })
func (m *Model[T]) Upsert(ctx context.Context, filter interface{}, update bson.M) (*mongo.UpdateResult, error) {
	injectUpdatedAt(update)
	// Add $setOnInsert for createdAt (only set on insert, not update)
	if _, ok := update["$setOnInsert"]; !ok {
		update["$setOnInsert"] = bson.M{"createdAt": time.Now()}
	} else if soi, ok := update["$setOnInsert"].(bson.M); ok {
		if _, has := soi["createdAt"]; !has {
			soi["createdAt"] = time.Now()
		}
	}
	opts := options.Update().SetUpsert(true)
	return m.Col().UpdateOne(ctx, filter, update, opts)
}
