package goose

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
)

// Model[T] is a typed wrapper around a MongoDB collection.
// Equivalent to: const MediaModel = mongoose.model('Media', mediaSchema)
type Model[T any] struct {
	collName string
	hooks    *hooks[T]
}

// NewModel creates a new Model for the given collection name.
//
//	var MediaModel = goose.NewModel[Media]("medias")
//	var FileModel  = goose.NewModel[File]("files")
func NewModel[T any](collection string) *Model[T] {
	return &Model[T]{collName: collection}
}

// Col returns the underlying mongo.Collection.
func (m *Model[T]) Col() *mongo.Collection {
	return db.Collection(m.collName)
}

// New creates a new document with defaults initialized.
// Reads `goose` struct tags for per-field defaults.
// Also calls BaseModel.InitDefaults() or SlugModel.InitDefaults() if embedded.
// Equivalent to: new MediaModel({})
func (m *Model[T]) New() *T {
	doc := new(T)
	// Apply goose struct tag defaults (uuid, random, now, literals)
	applyDefaults(doc)
	// Init SlugModel first (it includes BaseModel)
	if slug := getSlugModel(doc); slug != nil {
		slug.InitDefaults()
	} else if base := getBaseModel(doc); base != nil {
		base.InitDefaults()
	}
	return doc
}

// getBaseModel returns a pointer to the embedded BaseModel if present (direct or nested).
func getBaseModel(doc interface{}) *BaseModel {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	baseType := reflect.TypeOf(BaseModel{})
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == baseType && field.CanAddr() {
			return field.Addr().Interface().(*BaseModel)
		}
		// Check nested struct (e.g. SlugModel embeds BaseModel)
		if field.Kind() == reflect.Struct && field.CanAddr() {
			for j := 0; j < field.NumField(); j++ {
				inner := field.Field(j)
				if inner.Type() == baseType && inner.CanAddr() {
					return inner.Addr().Interface().(*BaseModel)
				}
			}
		}
	}
	return nil
}

// getSlugModel returns a pointer to the embedded SlugModel if present.
func getSlugModel(doc interface{}) *SlugModel {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	slugType := reflect.TypeOf(SlugModel{})
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == slugType && field.CanAddr() {
			return field.Addr().Interface().(*SlugModel)
		}
	}
	return nil
}

