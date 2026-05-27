package goose

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// ── Populate (Reference Join) ────────────────────────────────
//
// Populates reference fields by looking up documents from other collections.
// Similar to Mongoose's .populate() but uses reflection on goose struct tags.
//
// Usage:
//
//	type Post struct {
//	    goose.BaseModel `bson:",inline"`
//	    Title    string `bson:"title"    json:"title"`
//	    AuthorID string `bson:"authorId" json:"authorId" goose:"ref:users"`
//	}
//
//	// Populate a single ref field
//	author, err := goose.PopulateOne[User](ctx, "users", post.AuthorID)
//
//	// Populate multiple IDs at once
//	authors, err := goose.PopulateMany[User](ctx, "users", []string{id1, id2, id3})

// PopulateOne fetches a single referenced document by _id from the given collection.
//
//	user, err := goose.PopulateOne[User](ctx, "users", post.AuthorID)
func PopulateOne[T any](ctx context.Context, collection string, id string) (*T, error) {
	if id == "" {
		return nil, nil
	}
	doc := new(T)
	err := db.Collection(collection).FindOne(ctx, bson.M{"_id": id}).Decode(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// PopulateMany fetches multiple referenced documents by _id from the given collection.
// Returns a map of _id → document for easy lookup.
//
//	users, err := goose.PopulateMany[User](ctx, "users", []string{id1, id2})
//	author := users[post.AuthorID]
func PopulateMany[T any](ctx context.Context, collection string, ids []string) (map[string]*T, error) {
	if len(ids) == 0 {
		return map[string]*T{}, nil
	}

	// Deduplicate IDs
	unique := make(map[string]bool)
	var dedupIDs []string
	for _, id := range ids {
		if id != "" && !unique[id] {
			unique[id] = true
			dedupIDs = append(dedupIDs, id)
		}
	}

	cursor, err := db.Collection(collection).Find(ctx, bson.M{
		"_id": bson.M{"$in": dedupIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*T
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	// Build map by _id
	result := make(map[string]*T, len(docs))
	for _, doc := range docs {
		id := extractID(doc)
		if id != "" {
			result[id] = doc
		}
	}
	return result, nil
}

// PopulateField reads goose:"ref:xxx" tags from the document and populates
// matching fields automatically. The target field must be a pointer type with
// bson:"-" tag and goose:"populate:sourceField" tag.
//
// Example struct:
//
//	type Post struct {
//	    goose.BaseModel `bson:",inline"`
//	    AuthorID string `bson:"authorId" goose:"ref:users"`
//	    Author   *User  `bson:"-"        goose:"populate:authorId"`
//	}
//
//	goose.PopulateField(ctx, post, "Author")
func PopulateField(ctx context.Context, doc interface{}, fieldNames ...string) error {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("goose: PopulateField expects a struct, got %s", v.Kind())
	}
	t := v.Type()

	for _, targetFieldName := range fieldNames {
		targetField, ok := t.FieldByName(targetFieldName)
		if !ok {
			return fmt.Errorf("goose: field '%s' not found", targetFieldName)
		}

		// Parse goose:"populate:sourceField" tag
		gooseTag := targetField.Tag.Get("goose")
		var sourceFieldBson string
		for _, part := range strings.Split(gooseTag, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "populate:") {
				sourceFieldBson = part[len("populate:"):]
			}
		}
		if sourceFieldBson == "" {
			return fmt.Errorf("goose: field '%s' missing goose:\"populate:xxx\" tag", targetFieldName)
		}

		// Find the source field (by bson name) and its ref collection
		var sourceValue string
		var refCollection string
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			bsonTag := sf.Tag.Get("bson")
			bsonName := strings.Split(bsonTag, ",")[0]
			if bsonName == sourceFieldBson {
				sourceValue = v.Field(i).String()
				// Get ref from goose tag
				gooseTag := sf.Tag.Get("goose")
				for _, part := range strings.Split(gooseTag, ",") {
					part = strings.TrimSpace(part)
					if strings.HasPrefix(part, "ref:") {
						refCollection = part[len("ref:"):]
					}
				}
				break
			}
		}

		if refCollection == "" {
			return fmt.Errorf("goose: source field '%s' has no goose:\"ref:xxx\" tag", sourceFieldBson)
		}
		if sourceValue == "" {
			continue // skip empty refs
		}

		// Fetch the referenced document
		targetFieldValue := v.FieldByName(targetFieldName)
		if !targetFieldValue.CanSet() {
			return fmt.Errorf("goose: field '%s' is not settable", targetFieldName)
		}

		elemType := targetFieldValue.Type().Elem()
		newDoc := reflect.New(elemType)
		err := db.Collection(refCollection).FindOne(ctx, bson.M{"_id": sourceValue}).Decode(newDoc.Interface())
		if err != nil {
			return fmt.Errorf("goose: populate '%s' failed: %w", targetFieldName, err)
		}
		targetFieldValue.Set(newDoc)
	}

	return nil
}

// extractID gets _id from a document using reflection.
func extractID(doc interface{}) string {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		bsonTag := sf.Tag.Get("bson")
		bsonName := strings.Split(bsonTag, ",")[0]
		if bsonName == "_id" && v.Field(i).Kind() == reflect.String {
			return v.Field(i).String()
		}
		// Check embedded BaseModel
		if sf.Anonymous && sf.Type == reflect.TypeOf(BaseModel{}) {
			base := v.Field(i).Interface().(BaseModel)
			return base.ID
		}
	}
	return ""
}
