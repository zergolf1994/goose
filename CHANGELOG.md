# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-05-27

### Added

#### Middleware / Hooks (`hooks.go`) — P1
- `Pre(event, fn)` — register pre-hook for document operations (create)
- `Post(event, fn)` — register post-hook for document operations (create, find)
- `PreFilter(event, fn)` — register pre-hook for filter operations (update, delete, find)
- `PostFilter(event, fn)` — register post-hook for filter operations
- Supported events: `"create"`, `"update"`, `"delete"`, `"find"`
- Thread-safe hook registration with `sync.RWMutex`

#### Populate / Reference Join (`populate.go`) — P2
- `PopulateOne[T](ctx, collection, id)` — fetch a single referenced document
- `PopulateMany[T](ctx, collection, ids)` — batch fetch with dedup, returns `map[string]*T`
- `PopulateField(ctx, doc, fieldNames...)` — auto-populate using `goose:"populate:xxx"` struct tags

#### Multiple Connections (`connection.go`) — P2
- `CreateConnection(uri)` — create a separate connection (like `mongoose.createConnection()`)
- `CreateConnectionWithDB(db)` — wrap an existing `*mongo.Database`
- `NewConnModel[T](conn, collection)` — create a typed model bound to a specific connection
- `Connection.Client()`, `Connection.DB()`, `Connection.Collection()`, `Connection.Close()`, `Connection.Ping()`

#### Soft Delete (`softdelete.go`) — P3
- `SoftDeleteModel` — embeddable struct with `deletedAt` field
- `SoftDelete(ctx, filter)` / `SoftDeleteByID(ctx, id)` / `SoftDeleteMany(ctx, filter)`
- `Restore(ctx, filter)` / `RestoreByID(ctx, id)` / `RestoreMany(ctx, filter)`
- `FindActive(ctx, filter)` / `FindOneActive(ctx, filter)` — exclude soft-deleted
- `FindDeleted(ctx, filter)` — find only soft-deleted
- `CountActive(ctx, filter)` / `CountDeleted(ctx, filter)` / `ExistsActive(ctx, filter)`
- `QueryActive(filter)` — chainable query excluding soft-deleted
- `PermanentDelete(ctx, filter)` / `PermanentDeleteMany(ctx, filter)` — true hard delete

### Changed
- `Model[T]` struct now includes `hooks` field for middleware support
- All CRUD methods now run pre/post hooks:
  - `Create`, `Save`, `InsertMany` → pre/post `"create"` doc hooks
  - `FindOne`, `Find` → pre/post `"find"` filter+doc hooks
  - `UpdateOne`, `UpdateMany` → pre/post `"update"` filter hooks
  - `DeleteOne`, `DeleteMany` → pre/post `"delete"` filter hooks
- `CreateWithoutValidation` now also runs hooks (skips validation only)

---

## [0.1.0] - 2025-05-27

### Added

#### FindOneAndUpdate / FindOneAndDelete (`findandmodify.go`) — P0
- `FindOneAndUpdate(ctx, filter, update)` — atomically find and update, returns updated doc
- `FindByIDAndUpdate(ctx, id, update)` — shorthand for FindOneAndUpdate by `_id`
- `FindOneAndDelete(ctx, filter)` — atomically find and delete, returns deleted doc
- `FindByIDAndDelete(ctx, id)` — shorthand for FindOneAndDelete by `_id`
- `FindOneAndReplace(ctx, filter, replacement)` — atomically find and replace entire doc
- `Upsert(ctx, filter, update)` — update if exists, insert if not (auto `$setOnInsert` for `createdAt`)

#### Validation Engine (`validate.go`) — P0
- `Validate(doc)` — validates a document against goose struct tags, returns `*ValidationErrors`
- `ValidateField(doc, fieldName)` — validates a single field
- **Supported validators:**
  - `required` — field must not be zero-value (now enforced, not just metadata)
  - `enum:val1|val2|val3` — string must be one of the allowed values
  - `min:N` — minimum value for numbers, minimum length for strings
  - `max:N` — maximum value for numbers, maximum length for strings
  - `minlength:N` — minimum string length
  - `maxlength:N` — maximum string length
  - `match:^pattern$` — string must match regex pattern
- Validation is **auto-called** in `Create()`, `Save()`, and `InsertMany()`
- `CreateWithoutValidation()` — bypass validation when needed

#### Transaction Helpers (`transaction.go`) — P0
- `WithTransaction(ctx, fn)` — run a function in a transaction with auto commit/abort
- `WithTransactionResult(ctx, fn)` — same but returns a result value
- `RunInSession(ctx, fn)` — run in a session without explicit transaction
- `StartSession()` — create a session for manual transaction control
- `Ping(ctx)` — check MongoDB connectivity

#### Bulk Operations (`bulk.go`) — P1
- `Distinct(ctx, fieldName, filter)` — get distinct values for a field
- `BulkWrite(ctx, models)` — execute multiple write operations in one call
- `EstimatedDocumentCount(ctx)` — fast approximate count for large collections

### Changed
- `FieldSchema` struct now includes `Enum`, `Min`, `Max`, `MinLength`, `MaxLength`, `Match` fields
- `DescribeSchema()` now prints all validation metadata
- `Create()` and `InsertMany()` now run validation before insert

---

## [0.0.2] - 2025-05-27

### Added
- `LICENSE` file (MIT) — required for pkg.go.dev documentation display

---

## [0.0.1] - 2025-05-27

### Added
- Initial release
- Connection management (`Connect`, `Close`, `SetDB`, `Client`, `DB`, `Collection`)
- `BaseModel` with auto `_id` (UUID), `slug` (random), `createdAt`, `updatedAt`
- Generic `Model[T]` with `NewModel[T](collection)` and `New()`
- Struct tag schema engine (`goose:"default:uuid"`, `goose:"default:random(11)"`, etc.)
- CRUD operations: `Create`, `Save`, `InsertMany`, `FindOne`, `FindByID`, `FindBySlug`, `Find`, `FindRaw`
- Update operations: `UpdateOne`, `UpdateByID`, `UpdateMany`, `UpdateOneRaw` (auto `updatedAt`)
- Delete operations: `DeleteOne`, `DeleteByID`, `DeleteMany`
- Count & utility: `CountDocuments`, `Exists`, `Aggregate`
- Chainable query builder: `Query()`, `Sort()`, `SortAsc()`, `SortDesc()`, `Limit()`, `Skip()`, `Page()`, `Select()`, `Exclude()`, `Exec()`, `One()`, `Count()`
- Auto index management: `EnsureIndexes()`, `EnsureCompoundIndex()`
- Schema inspection: `GetSchema[T]()`, `DescribeSchema[T]()`
