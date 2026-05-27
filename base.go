package goose

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// BaseModel provides auto-generated _id and timestamps.
// Embed this in your model structs with `bson:",inline"`.
//
// Fields: _id (uuid), createdAt, updatedAt
type BaseModel struct {
	ID        string    `bson:"_id" json:"id"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// InitDefaults sets _id, createdAt, updatedAt if not already set.
// Called automatically by Model.Create() and Model.Save().
func (b *BaseModel) InitDefaults() {
	now := time.Now()
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
}

// TouchUpdatedAt sets updatedAt to now.
func (b *BaseModel) TouchUpdatedAt() {
	b.UpdatedAt = time.Now()
}

// SlugModel extends BaseModel with a random slug field.
// Use this for models that need a short public-facing identifier.
//
// Fields: _id (uuid), slug (random 11-char), createdAt, updatedAt
type SlugModel struct {
	BaseModel `bson:",inline"`
	Slug      string `bson:"slug" json:"slug"`
}

// InitDefaults sets _id, slug, createdAt, updatedAt if not already set.
func (s *SlugModel) InitDefaults() {
	s.BaseModel.InitDefaults()
	if s.Slug == "" {
		s.Slug = randomSlug(11)
	}
}

// ── random slug generator ──

const alphanumChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// randomSlug generates a slug with dash and underscore (matches randomString from TS).
func randomSlug(length int) string {
	if length < 3 {
		return randomAlphaNum(length)
	}
	base := []byte(randomAlphaNum(length))

	dashPos := rand.Intn(length-2) + 1
	underscorePos := rand.Intn(length-2) + 1
	for dashPos == underscorePos {
		underscorePos = rand.Intn(length-2) + 1
	}

	result := make([]byte, 0, length+2)
	for i, c := range base {
		if i == dashPos {
			result = append(result, '-')
		}
		if i == underscorePos {
			result = append(result, '_')
		}
		result = append(result, c)
	}
	return string(result)
}

func randomAlphaNum(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphanumChars[rand.Intn(len(alphanumChars))]
	}
	return string(b)
}
