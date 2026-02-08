package ui

import "github.com/hajimehoshi/ebiten/v2"

// imagePool manages reusable offscreen ebiten.Image instances.
// It reduces per-frame GPU allocation overhead by recycling images in
// power-of-two size buckets.  All access is from the single Ebiten game loop
// goroutine — no mutex is needed.
type imagePool struct {
	// Key: [bucketWidth, bucketHeight], Value: available images of that bucket size.
	pool map[[2]int][]*ebiten.Image
}

// globalImagePool is the process-wide image pool used by all rendering code.
var globalImagePool = &imagePool{
	pool: make(map[[2]int][]*ebiten.Image),
}

// maxPerBucket caps the number of images retained per size bucket to prevent
// unbounded memory growth.  4 is sufficient for the deepest nesting of
// offscreen passes observed in practice (compositing → filter → backdrop).
const maxPerBucket = 4

// bucketSize rounds n up to the next power of two, with a minimum of 16.
// This limits the number of distinct allocation sizes and maximises reuse.
func bucketSize(n int) int {
	if n <= 0 {
		return 16
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	if n < 16 {
		n = 16
	}
	return n
}

// Get returns a cleared image of at least (w, h) pixels.
// The actual image may be larger due to bucket rounding.  Callers that use
// DrawRectShader or DrawImage with explicit (w, h) are unaffected by the extra
// transparent padding.
func (p *imagePool) Get(w, h int) *ebiten.Image {
	bw, bh := bucketSize(w), bucketSize(h)
	key := [2]int{bw, bh}

	if imgs, ok := p.pool[key]; ok && len(imgs) > 0 {
		img := imgs[len(imgs)-1]
		p.pool[key] = imgs[:len(imgs)-1]
		img.Clear()
		return img
	}

	return ebiten.NewImage(bw, bh)
}

// Put returns an image to the pool for later reuse.
// The caller must NOT use the image after this call.  If the bucket is already
// at capacity the image is deallocated to bound memory.
func (p *imagePool) Put(img *ebiten.Image) {
	if img == nil {
		return
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	key := [2]int{w, h}

	if imgs, ok := p.pool[key]; ok && len(imgs) >= maxPerBucket {
		img.Deallocate()
		return
	}

	p.pool[key] = append(p.pool[key], img)
}
