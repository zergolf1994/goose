package goose

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ── Multiple Connections ─────────────────────────────────────
//
// Supports multiple database connections beyond the default global one.
// Similar to Mongoose's mongoose.createConnection().
//
// Usage:
//
//	conn, err := goose.CreateConnection("mongodb://host2:27017/analytics")
//	defer conn.Close()
//
//	var LogModel = conn.NewModel[Log]("logs")
//	LogModel.Create(ctx, log)

// Connection represents a separate database connection.
type Connection struct {
	client *mongo.Client
	db     *mongo.Database
	uri    string
}

// CreateConnection creates a new connection to a different MongoDB instance/database.
// Equivalent to: mongoose.createConnection(uri)
func CreateConnection(uri string) (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := c.Ping(ctx, nil); err != nil {
		return nil, err
	}

	dbName := parseDBName(uri)
	conn := &Connection{
		client: c,
		db:     c.Database(dbName),
		uri:    uri,
	}
	log.Printf("✅ goose: new connection to MongoDB: %s", dbName)
	return conn, nil
}

// CreateConnectionWithDB creates a connection with a specific database from an existing client.
func CreateConnectionWithDB(database *mongo.Database) *Connection {
	return &Connection{
		db: database,
	}
}

// Client returns the underlying mongo.Client.
func (conn *Connection) Client() *mongo.Client {
	return conn.client
}

// DB returns the database instance.
func (conn *Connection) DB() *mongo.Database {
	return conn.db
}

// Collection returns a raw mongo.Collection by name.
func (conn *Connection) Collection(name string) *mongo.Collection {
	return conn.db.Collection(name)
}

// Close disconnects this connection.
func (conn *Connection) Close() error {
	if conn.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return conn.client.Disconnect(ctx)
	}
	return nil
}

// Ping checks connectivity for this connection.
func (conn *Connection) Ping(ctx context.Context) error {
	if conn.client == nil {
		return mongo.ErrClientDisconnected
	}
	return conn.client.Ping(ctx, nil)
}

// ── ConnModel: Model bound to a specific Connection ─────────

// ConnModel[T] is a Model that uses a specific connection instead of the global DB.
type ConnModel[T any] struct {
	Model[T]
	conn *Connection
}

// NewModel creates a new Model bound to this connection.
//
//	conn, _ := goose.CreateConnection("mongodb://host2:27017/analytics")
//	var LogModel = conn.NewModel[Log]("logs")
func (conn *Connection) NewModel(collection string) *ConnModel[struct{}] {
	return &ConnModel[struct{}]{
		Model: Model[struct{}]{collName: collection},
		conn:  conn,
	}
}

// Col returns the collection from this connection (not the global DB).
func (m *ConnModel[T]) Col() *mongo.Collection {
	return m.conn.db.Collection(m.collName)
}

// NewConnModel creates a typed model bound to a specific connection.
// Use this generic function instead of conn.NewModel for full type safety.
//
//	var LogModel = goose.NewConnModel[Log](conn, "logs")
func NewConnModel[T any](conn *Connection, collection string) *ConnModel[T] {
	return &ConnModel[T]{
		Model: Model[T]{collName: collection},
		conn:  conn,
	}
}
