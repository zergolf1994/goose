package goose

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// WithTransaction runs a function within a MongoDB transaction.
// Automatically handles session start, commit, and abort on error.
// The callback receives a mongo.SessionContext that should be passed to all DB operations.
//
// Equivalent to:
//
//	const session = await mongoose.startSession();
//	await session.withTransaction(async () => { ... });
//
// Usage:
//
//	err := goose.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
//	    UserModel.Create(sc, user)
//	    PostModel.DeleteMany(sc, bson.M{"userId": id})
//	    return nil, nil
//	})
func WithTransaction(ctx context.Context, fn func(sc mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) error {
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, fn, opts...)
	return err
}

// WithTransactionResult runs a function within a transaction and returns the result.
// Use this when you need the return value from the transaction.
func WithTransactionResult(ctx context.Context, fn func(sc mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error) {
	session, err := client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, fn, opts...)
}

// RunInSession runs a function within a session without an explicit transaction.
// Useful for causal consistency without full transaction overhead.
func RunInSession(ctx context.Context, fn func(sc mongo.SessionContext) error) error {
	return client.UseSession(ctx, fn)
}

// StartSession creates a new session for manual transaction control.
// Remember to call session.EndSession(ctx) when done.
//
// Usage:
//
//	session, _ := goose.StartSession()
//	defer session.EndSession(ctx)
//	session.StartTransaction()
//	// ... do operations ...
//	session.CommitTransaction(ctx)
func StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return client.StartSession(opts...)
}

// Ping checks connectivity to MongoDB.
func Ping(ctx context.Context) error {
	if client == nil {
		return mongo.ErrClientDisconnected
	}
	return client.Ping(ctx, readpref.Primary())
}
