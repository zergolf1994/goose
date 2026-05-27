# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-05-27

### Added

#### FindOneAndUpdate / FindOneAndDelete (`findandmodify.go`)
- `FindOneAndUpdate(ctx, filter, update)` — atomically find and update, returns updated doc
- `FindByIDAndUpdate(ctx, id, update)` — shorthand for FindOneAndUpdate by `_id`
- `FindOneAndDelete(ctx, filter)` — atomically find and delete, returns deleted doc
- `FindByIDAndDelete(ctx, id)` — shorthand for FindOneAndDelete by `_id`
- `FindOneAndReplace(ctx, filter, replacement)` — atomically find and replace entire doc
- `Upsert(ctx, filter, update)` — update if exists, insert if not (auto `$setOnInsert` for `createdAt`)

#### Validation Engine (`validate.go`)
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

#### Transaction Helpers (`transaction.go`)
- `WithTransaction(ctx, fn)` — run a function in a transaction with auto commit/abort
- `WithTransactionResult(ctx, fn)` — same but returns a result value
- `RunInSession(ctx, fn)` — run in a session without explicit transaction
- `StartSession()` — create a session for manual transaction control
- `Ping(ctx)` — check MongoDB connectivity

#### Bulk Operations (`bulk.go`)
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
