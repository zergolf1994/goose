package goose

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Query[T] is a chainable query builder for MongoDB find operations.
//
// Usage:
//
//	results, err := FileModel.Query(bson.M{"type": "video"}).
//	    Sort("createdAt", -1).
//	    Limit(10).
//	    Skip(20).
//	    Select("name", "status", "createdAt").
//	    Exec(ctx)
//
//	// Single result
//	file, err := FileModel.Query(bson.M{"slug": "abc"}).One(ctx)
type Query[T any] struct {
	model      *Model[T]
	filter     interface{}
	sort       bson.D
	limit      *int64
	skip       *int64
	projection bson.M
}

// Query starts a chainable query with the given filter.
func (m *Model[T]) Query(filter interface{}) *Query[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return &Query[T]{
		model:  m,
		filter: filter,
	}
}

// Sort adds a sort field. Use 1 for ascending, -1 for descending.
func (q *Query[T]) Sort(field string, order int) *Query[T] {
	q.sort = append(q.sort, bson.E{Key: field, Value: order})
	return q
}

// SortDesc is a shorthand for .Sort(field, -1)
func (q *Query[T]) SortDesc(field string) *Query[T] {
	return q.Sort(field, -1)
}

// SortAsc is a shorthand for .Sort(field, 1)
func (q *Query[T]) SortAsc(field string) *Query[T] {
	return q.Sort(field, 1)
}

// Limit sets the maximum number of results to return.
func (q *Query[T]) Limit(n int64) *Query[T] {
	q.limit = &n
	return q
}

// Skip sets the number of results to skip (for pagination).
func (q *Query[T]) Skip(n int64) *Query[T] {
	q.skip = &n
	return q
}

// Select specifies which fields to include in the result (projection).
func (q *Query[T]) Select(fields ...string) *Query[T] {
	if q.projection == nil {
		q.projection = bson.M{}
	}
	for _, f := range fields {
		q.projection[f] = 1
	}
	return q
}

// Exclude specifies which fields to exclude from the result.
func (q *Query[T]) Exclude(fields ...string) *Query[T] {
	if q.projection == nil {
		q.projection = bson.M{}
	}
	for _, f := range fields {
		q.projection[f] = 0
	}
	return q
}

// Page is a pagination helper. Sets skip and limit from page number and page size.
// Page is 1-indexed (page 1 = first page).
func (q *Query[T]) Page(page, pageSize int64) *Query[T] {
	if page < 1 {
		page = 1
	}
	skip := (page - 1) * pageSize
	q.skip = &skip
	q.limit = &pageSize
	return q
}

// Exec executes the query and returns all matching documents.
func (q *Query[T]) Exec(ctx context.Context) ([]*T, error) {
	opts := options.Find()

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}
	if q.limit != nil {
		opts.SetLimit(*q.limit)
	}
	if q.skip != nil {
		opts.SetSkip(*q.skip)
	}
	if len(q.projection) > 0 {
		opts.SetProjection(q.projection)
	}

	return q.model.Find(ctx, q.filter, opts)
}

// One executes the query and returns the first matching document.
func (q *Query[T]) One(ctx context.Context) (*T, error) {
	opts := options.FindOne()

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}
	if q.skip != nil {
		opts.SetSkip(*q.skip)
	}
	if len(q.projection) > 0 {
		opts.SetProjection(q.projection)
	}

	return q.model.FindOne(ctx, q.filter, opts)
}

// Count returns the number of documents matching the query filter.
func (q *Query[T]) Count(ctx context.Context) (int64, error) {
	return q.model.CountDocuments(ctx, q.filter)
}
