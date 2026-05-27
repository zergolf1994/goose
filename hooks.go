package goose

import (
	"context"
	"sync"
)

// ── Middleware / Hooks Engine ─────────────────────────────────
//
// Provides Mongoose-like pre/post hooks for CRUD operations.
//
// Supported hook events:
//   - "create"   → before/after Create, Save
//   - "update"   → before/after UpdateOne, UpdateByID, UpdateMany
//   - "delete"   → before/after DeleteOne, DeleteByID, DeleteMany
//   - "find"     → before/after FindOne, FindByID, Find
//
// Usage:
//
//	MediaModel.Pre("create", func(ctx context.Context, doc *Media) error {
//	    doc.Status = strings.ToLower(doc.Status)
//	    return nil
//	})
//
//	MediaModel.Post("create", func(ctx context.Context, doc *Media) error {
//	    log.Printf("Created media: %s", doc.ID)
//	    return nil
//	})

// HookType represents when a hook runs.
type HookType string

const (
	HookPre  HookType = "pre"
	HookPost HookType = "post"
)

// HookEvent represents which operation triggers the hook.
type HookEvent string

const (
	EventCreate HookEvent = "create"
	EventUpdate HookEvent = "update"
	EventDelete HookEvent = "delete"
	EventFind   HookEvent = "find"
)

// hookKey uniquely identifies a hook registration.
type hookKey struct {
	hookType HookType
	event    HookEvent
}

// DocHookFunc is a hook function that receives the document being operated on.
type DocHookFunc[T any] func(ctx context.Context, doc *T) error

// FilterHookFunc is a hook function that receives the filter used in an operation.
type FilterHookFunc func(ctx context.Context, filter interface{}) error

// hooks stores registered hook functions for a model.
type hooks[T any] struct {
	mu          sync.RWMutex
	docHooks    map[hookKey][]DocHookFunc[T]
	filterHooks map[hookKey][]FilterHookFunc
}

func newHooks[T any]() *hooks[T] {
	return &hooks[T]{
		docHooks:    make(map[hookKey][]DocHookFunc[T]),
		filterHooks: make(map[hookKey][]FilterHookFunc),
	}
}

// ── Model hook registration ─────────────────────────────────

// Pre registers a pre-hook for document operations (create).
// The hook receives the document before it's saved to the database.
//
//	MediaModel.Pre("create", func(ctx context.Context, doc *Media) error {
//	    doc.Status = strings.ToLower(doc.Status)
//	    return nil
//	})
func (m *Model[T]) Pre(event string, fn DocHookFunc[T]) {
	m.initHooks()
	key := hookKey{HookPre, HookEvent(event)}
	m.hooks.mu.Lock()
	defer m.hooks.mu.Unlock()
	m.hooks.docHooks[key] = append(m.hooks.docHooks[key], fn)
}

// Post registers a post-hook for document operations (create, find).
// The hook receives the document after the database operation completes.
//
//	MediaModel.Post("create", func(ctx context.Context, doc *Media) error {
//	    log.Printf("Created: %s", doc.ID)
//	    return nil
//	})
func (m *Model[T]) Post(event string, fn DocHookFunc[T]) {
	m.initHooks()
	key := hookKey{HookPost, HookEvent(event)}
	m.hooks.mu.Lock()
	defer m.hooks.mu.Unlock()
	m.hooks.docHooks[key] = append(m.hooks.docHooks[key], fn)
}

// PreFilter registers a pre-hook for filter-based operations (update, delete, find).
// The hook receives the filter before the operation executes.
//
//	MediaModel.PreFilter("delete", func(ctx context.Context, filter interface{}) error {
//	    log.Printf("Deleting with filter: %v", filter)
//	    return nil
//	})
func (m *Model[T]) PreFilter(event string, fn FilterHookFunc) {
	m.initHooks()
	key := hookKey{HookPre, HookEvent(event)}
	m.hooks.mu.Lock()
	defer m.hooks.mu.Unlock()
	m.hooks.filterHooks[key] = append(m.hooks.filterHooks[key], fn)
}

// PostFilter registers a post-hook for filter-based operations.
func (m *Model[T]) PostFilter(event string, fn FilterHookFunc) {
	m.initHooks()
	key := hookKey{HookPost, HookEvent(event)}
	m.hooks.mu.Lock()
	defer m.hooks.mu.Unlock()
	m.hooks.filterHooks[key] = append(m.hooks.filterHooks[key], fn)
}

// ── Internal hook execution ─────────────────────────────────

func (m *Model[T]) initHooks() {
	if m.hooks == nil {
		m.hooks = newHooks[T]()
	}
}

// runDocHooks executes all registered document hooks for the given type+event.
func (m *Model[T]) runDocHooks(ctx context.Context, ht HookType, event HookEvent, doc *T) error {
	if m.hooks == nil {
		return nil
	}
	key := hookKey{ht, event}
	m.hooks.mu.RLock()
	fns := m.hooks.docHooks[key]
	m.hooks.mu.RUnlock()

	for _, fn := range fns {
		if err := fn(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// runDocHooksMany executes document hooks for multiple documents.
func (m *Model[T]) runDocHooksMany(ctx context.Context, ht HookType, event HookEvent, docs []*T) error {
	if m.hooks == nil {
		return nil
	}
	for _, doc := range docs {
		if err := m.runDocHooks(ctx, ht, event, doc); err != nil {
			return err
		}
	}
	return nil
}

// runFilterHooks executes all registered filter hooks for the given type+event.
func (m *Model[T]) runFilterHooks(ctx context.Context, ht HookType, event HookEvent, filter interface{}) error {
	if m.hooks == nil {
		return nil
	}
	key := hookKey{ht, event}
	m.hooks.mu.RLock()
	fns := m.hooks.filterHooks[key]
	m.hooks.mu.RUnlock()

	for _, fn := range fns {
		if err := fn(ctx, filter); err != nil {
			return err
		}
	}
	return nil
}
