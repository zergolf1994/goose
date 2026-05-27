package goose

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Create inserts a new document, auto-setting defaults and validating before insert.
// Runs pre("create") hooks → defaults → validate → insert → post("create") hooks.
// Equivalent to: await MediaModel.create({...})
func (m *Model[T]) Create(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	// Pre-create hooks
	if err := m.runDocHooks(ctx, HookPre, EventCreate, doc); err != nil {
		return nil, err
	}
	// Apply goose struct tag defaults
	applyDefaults(doc)
	// Also init BaseModel if embedded
	if base := getBaseModel(doc); base != nil {
		base.InitDefaults()
	}
	// Validate before insert
	if err := Validate(doc); err != nil {
		return nil, err
	}
	result, err := m.Col().InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	// Post-create hooks
	_ = m.runDocHooks(ctx, HookPost, EventCreate, doc)
	return result, nil
}

// CreateWithoutValidation inserts a new document without running validators.
// Still runs hooks. Use when you've already validated or need to bypass validation.
func (m *Model[T]) CreateWithoutValidation(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	if err := m.runDocHooks(ctx, HookPre, EventCreate, doc); err != nil {
		return nil, err
	}
	applyDefaults(doc)
	if base := getBaseModel(doc); base != nil {
		base.InitDefaults()
	}
	result, err := m.Col().InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	_ = m.runDocHooks(ctx, HookPost, EventCreate, doc)
	return result, nil
}

// Save is an alias for Create (insert only).
// Equivalent to: await doc.save()
func (m *Model[T]) Save(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	return m.Create(ctx, doc)
}

// InsertMany inserts multiple documents, auto-setting defaults and validating.
// Runs pre("create") hooks on each document.
// Equivalent to: await MediaModel.insertMany([...])
func (m *Model[T]) InsertMany(ctx context.Context, docs []*T) (*mongo.InsertManyResult, error) {
	iDocs := make([]interface{}, len(docs))
	for i, doc := range docs {
		if err := m.runDocHooks(ctx, HookPre, EventCreate, doc); err != nil {
			return nil, err
		}
		applyDefaults(doc)
		if base := getBaseModel(doc); base != nil {
			base.InitDefaults()
		}
		if err := Validate(doc); err != nil {
			return nil, err
		}
		iDocs[i] = doc
	}
	result, err := m.Col().InsertMany(ctx, iDocs)
	if err != nil {
		return nil, err
	}
	// Post hooks for each doc
	_ = m.runDocHooksMany(ctx, HookPost, EventCreate, docs)
	return result, nil
}
