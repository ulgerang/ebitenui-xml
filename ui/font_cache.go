package ui

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// fontCacheKey identifies a unique font face configuration.
type fontCacheKey struct {
	size float64
}

// FontCache caches text.GoTextFace instances keyed by size to avoid redundant
// allocations when many widgets share the same font source. Thread-safe via
// sync.RWMutex for concurrent read access during the rendering hot path.
type FontCache struct {
	source *text.GoTextFaceSource
	mu     sync.RWMutex
	faces  map[fontCacheKey]*text.GoTextFace
}

// NewFontCache creates a new FontCache backed by the given GoTextFaceSource.
func NewFontCache(source *text.GoTextFaceSource) *FontCache {
	return &FontCache{
		source: source,
		faces:  make(map[fontCacheKey]*text.GoTextFace),
	}
}

// GetFace returns a cached GoTextFace for the given size, creating one if
// needed. Uses double-checked locking to minimise write-lock contention.
func (fc *FontCache) GetFace(size float64) *text.GoTextFace {
	if size <= 0 {
		size = 14
	}

	key := fontCacheKey{size: size}

	// Fast path: read lock
	fc.mu.RLock()
	if face, ok := fc.faces[key]; ok {
		fc.mu.RUnlock()
		return face
	}
	fc.mu.RUnlock()

	// Slow path: write lock
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Double-check after acquiring write lock
	if face, ok := fc.faces[key]; ok {
		return face
	}

	face := &text.GoTextFace{
		Source: fc.source,
		Size:   size,
	}
	fc.faces[key] = face
	return face
}

// Source returns the underlying GoTextFaceSource.
func (fc *FontCache) Source() *text.GoTextFaceSource {
	return fc.source
}
