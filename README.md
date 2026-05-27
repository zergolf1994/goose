# 🪿 Goose

**A Mongoose-inspired ODM for Go + MongoDB**

Goose brings the familiar, elegant API of [Mongoose](https://mongoosejs.com/) to Go — with generics, struct tags, and zero boilerplate. If you've used Mongoose in Node.js, you'll feel right at home.

---

## ✨ Features

- **🔷 Generics-based models** — fully typed `Model[T]` with compile-time safety
- **🏷️ Struct tag schema** — define defaults, indexes, and refs with `goose:"..."` tags
- **🆔 Auto `_id` + `slug`** — UUID-based `_id` and random slug generated automatically
- **🕐 Auto timestamps** — `createdAt` / `updatedAt` managed for you
- **🔗 Chainable query builder** — `.Sort()`, `.Limit()`, `.Skip()`, `.Select()`, `.Page()`
- **📇 Auto indexes** — unique & compound indexes from struct tags
- **📦 Zero config** — just `Connect()` and go

---

## 📦 Installation

```bash
go get github.com/zergolf1994/goose
```

**Requirements:**
- Go 1.21+ (generics support)
- MongoDB 4.0+

---

## 🚀 Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/zergolf1994/goose"
    "go.mongodb.org/mongo-driver/bson"
)

// 1. Define your model
type Media struct {
    goose.BaseModel `bson:",inline"`
    Type   string   `bson:"type"   json:"type"   goose:"default:video"`
    FileID string   `bson:"fileId" json:"fileId"  goose:"required,index"`
}

// 2. Create model instance
var MediaModel = goose.NewModel[Media]("medias")

func main() {
    ctx := context.Background()

    // 3. Connect to MongoDB
    if err := goose.Connect("mongodb://localhost:27017/mydb"); err != nil {
        log.Fatal(err)
    }
    defer goose.Close()

    // 4. Auto-create indexes
    MediaModel.EnsureIndexes(ctx)

    // 5. Create a document
    media := MediaModel.New()
    media.Type = "video"
    media.FileID = "file_abc123"
    MediaModel.Create(ctx, media)

    // 6. Query documents
    results, _ := MediaModel.Query(bson.M{"type": "video"}).
        SortDesc("createdAt").
        Limit(10).
        Exec(ctx)

    log.Printf("Found %d videos", len(results))
}
```

---

## 📖 API Reference

### Connection

| Function | Description |
|---|---|
| `goose.Connect(uri)` | Connect to MongoDB (auto-parses DB name from URI) |
| `goose.SetDB(db)` | Use an existing `*mongo.Database` instance |
| `goose.Close()` | Disconnect from MongoDB |
| `goose.Client()` | Get the underlying `*mongo.Client` |
| `goose.DB()` | Get the current `*mongo.Database` |
| `goose.Collection(name)` | Get a raw `*mongo.Collection` |

### Model

| Method | Description |
|---|---|
| `goose.NewModel[T](collection)` | Create a typed model for a collection |
| `model.New()` | Create a new doc with defaults applied |
| `model.Col()` | Get the underlying `*mongo.Collection` |

### CRUD Operations

| Method | Mongoose Equivalent |
|---|---|
| `model.Create(ctx, doc)` | `Model.create({...})` |
| `model.Save(ctx, doc)` | `doc.save()` |
| `model.InsertMany(ctx, docs)` | `Model.insertMany([...])` |
| `model.FindOne(ctx, filter)` | `Model.findOne({...})` |
| `model.FindByID(ctx, id)` | `Model.findById(id)` |
| `model.FindBySlug(ctx, slug)` | `Model.findOne({ slug })` |
| `model.Find(ctx, filter)` | `Model.find({...})` |
| `model.FindRaw(ctx, filter)` | Raw cursor for advanced use |
| `model.UpdateOne(ctx, filter, update)` | `Model.updateOne(filter, update)` |
| `model.UpdateByID(ctx, id, update)` | `Model.findByIdAndUpdate(id, update)` |
| `model.UpdateMany(ctx, filter, update)` | `Model.updateMany(filter, update)` |
| `model.UpdateOneRaw(ctx, filter, update)` | Update without auto `updatedAt` |
| `model.DeleteOne(ctx, filter)` | `Model.deleteOne(filter)` |
| `model.DeleteByID(ctx, id)` | `Model.findByIdAndDelete(id)` |
| `model.DeleteMany(ctx, filter)` | `Model.deleteMany(filter)` |
| `model.CountDocuments(ctx, filter)` | `Model.countDocuments(filter)` |
| `model.Exists(ctx, filter)` | Check if any doc matches |
| `model.Aggregate(ctx, pipeline)` | `Model.aggregate([...])` |

### Query Builder

```go
results, err := MediaModel.Query(bson.M{"type": "video"}).
    Sort("createdAt", -1).     // Sort by field (1 = asc, -1 = desc)
    SortDesc("createdAt").     // Shorthand for descending
    SortAsc("name").           // Shorthand for ascending
    Limit(10).                 // Limit results
    Skip(20).                  // Skip N results
    Page(2, 10).               // Pagination helper (page, pageSize)
    Select("name", "status").  // Include only specific fields
    Exclude("password").       // Exclude specific fields
    Exec(ctx)                  // Execute and return []*T

// Single result
doc, err := MediaModel.Query(bson.M{"slug": "abc"}).One(ctx)

// Count
count, err := MediaModel.Query(bson.M{"type": "video"}).Count(ctx)
```

### Indexes

```go
// Auto-create indexes from struct tags
MediaModel.EnsureIndexes(ctx)

// Compound index
MediaModel.EnsureCompoundIndex(ctx, []string{"type", "status"}, false)

// Unique compound index
MediaModel.EnsureCompoundIndex(ctx, []string{"userId", "email"}, true)
```

---

## 🏷️ Struct Tags

Goose uses the `goose` struct tag to define schema metadata:

```go
type File struct {
    goose.BaseModel `bson:",inline"`

    Name     string `bson:"name"     goose:"required"`
    Code     string `bson:"code"     goose:"default:uuid"`
    Slug     string `bson:"slug"     goose:"default:random(11),unique"`
    Status   string `bson:"status"   goose:"default:waiting,index"`
    FileID   string `bson:"fileId"   goose:"required,ref:files"`
    Priority int    `bson:"priority" goose:"default:0"`
    Active   bool   `bson:"active"   goose:"default:true"`
}
```

| Tag | Description | Example |
|---|---|---|
| `default:uuid` | Auto-generate UUID string | `goose:"default:uuid"` |
| `default:random(N)` | Random slug of length N | `goose:"default:random(11)"` |
| `default:now` | `time.Now()` for time fields | `goose:"default:now"` |
| `default:<value>` | Literal default value | `goose:"default:waiting"` |
| `required` | Mark field as required (metadata) | `goose:"required"` |
| `unique` | Create unique index | `goose:"unique"` |
| `index` | Create regular index | `goose:"index"` |
| `ref:<collection>` | Reference to another collection | `goose:"ref:files"` |

Tags can be combined with commas: `goose:"default:uuid,unique"`

---

## 🧬 BaseModel

Embed `goose.BaseModel` to get automatic fields:

```go
type BaseModel struct {
    ID        string    `bson:"_id"       json:"id"`         // UUID v4
    Slug      string    `bson:"slug"      json:"slug"`       // Random 11-char slug
    CreatedAt time.Time `bson:"createdAt" json:"createdAt"`  // Auto-set on create
    UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`  // Auto-set on create & update
}
```

These fields are automatically initialized when you call `model.New()` or `model.Create()`.

---

## 🔄 Auto `updatedAt`

`UpdateOne`, `UpdateByID`, and `UpdateMany` automatically inject `updatedAt` into the `$set` operator:

```go
// This automatically sets updatedAt = time.Now()
MediaModel.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
    "$set": bson.M{"type": "audio"},
})

// Use UpdateOneRaw to skip auto-injection
MediaModel.UpdateOneRaw(ctx, filter, bson.M{
    "$push": bson.M{"tags": "new-tag"},
})
```

---

## 📂 Project Structure

```
goose/
├── goose.go     # Connection management (Connect, Close, DB, Client)
├── base.go      # BaseModel with auto _id, slug, timestamps
├── model.go     # Generic Model[T] and NewModel constructor
├── schema.go    # Struct tag parser and default value engine
├── find.go      # FindOne, FindByID, FindBySlug, Find, FindRaw
├── insert.go    # Create, Save, InsertMany
├── update.go    # UpdateOne, UpdateByID, UpdateMany, UpdateOneRaw
├── delete.go    # DeleteOne, DeleteByID, DeleteMany
├── count.go     # CountDocuments, Exists, Aggregate
├── query.go     # Chainable query builder
└── index.go     # Auto index creation from struct tags
```

---

## 📄 License

MIT

---

## 🙏 Acknowledgments

Inspired by [Mongoose](https://mongoosejs.com/) — the elegant MongoDB ODM for Node.js.
