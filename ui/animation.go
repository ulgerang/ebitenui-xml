// Package ui - Animation System for Modern UI Effects
package ui

import (
	"image/color"
	"math"
	"time"
)

// ============================================================================
// Animation Types and Keyframes
// ============================================================================

// Animation represents a CSS-like animation
type Animation struct {
	Name           string
	Keyframes      []Keyframe
	Duration       time.Duration
	Delay          time.Duration
	IterationCount int // -1 for infinite
	Direction      AnimationDirection
	FillMode       AnimationFillMode
	TimingFunc     EasingFunc

	// Runtime state
	startTime    time.Time
	isPlaying    bool
	currentFrame float64 // 0-1 progress
	iteration    int
}

// Keyframe represents a single keyframe in an animation
type Keyframe struct {
	Percent    float64 // 0-100 (0% to 100%)
	Properties KeyframeProperties
}

// KeyframeProperties holds animatable properties at a keyframe
type KeyframeProperties struct {
	// Transform
	TranslateX float64
	TranslateY float64
	ScaleX     float64
	ScaleY     float64
	Rotate     float64 // degrees
	SkewX      float64
	SkewY      float64

	// Visual
	Opacity         float64
	BackgroundColor color.Color
	BorderColor     color.Color
	BoxShadowBlur   float64
	BoxShadowSpread float64

	// Size
	Width  float64
	Height float64

	// Position offset
	OffsetX float64
	OffsetY float64
}

type AnimationDirection int

const (
	AnimationNormal AnimationDirection = iota
	AnimationReverse
	AnimationAlternate
	AnimationAlternateReverse
)

type AnimationFillMode int

const (
	AnimationFillNone AnimationFillMode = iota
	AnimationFillForwards
	AnimationFillBackwards
	AnimationFillBoth
)

// ============================================================================
// Pre-built Animations (like animate.css)
// ============================================================================

// Built-in animation presets
var (
	// Fade animations
	AnimFadeIn = &Animation{
		Name:     "fadeIn",
		Duration: 300 * time.Millisecond,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{Opacity: 1}},
		},
	}

	AnimFadeOut = &Animation{
		Name:     "fadeOut",
		Duration: 300 * time.Millisecond,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{Opacity: 1}},
			{Percent: 100, Properties: KeyframeProperties{Opacity: 0}},
		},
	}

	// Pulse animation - 3 times then stop
	AnimPulse = &Animation{
		Name:           "pulse",
		Duration:       400 * time.Millisecond,
		IterationCount: 3, // 3 times
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1}},
			{Percent: 50, Properties: KeyframeProperties{ScaleX: 1.15, ScaleY: 1.15}},
			{Percent: 100, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1}},
		},
	}

	// Bounce animation - more dramatic bouncing
	AnimBounce = &Animation{
		Name:           "bounce",
		Duration:       800 * time.Millisecond,
		IterationCount: 1,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateY: 0, ScaleX: 1, ScaleY: 1}},
			{Percent: 20, Properties: KeyframeProperties{TranslateY: -25, ScaleX: 1.1, ScaleY: 0.9}},
			{Percent: 40, Properties: KeyframeProperties{TranslateY: 0, ScaleX: 0.95, ScaleY: 1.05}},
			{Percent: 60, Properties: KeyframeProperties{TranslateY: -12, ScaleX: 1.03, ScaleY: 0.97}},
			{Percent: 80, Properties: KeyframeProperties{TranslateY: 0, ScaleX: 0.98, ScaleY: 1.02}},
			{Percent: 100, Properties: KeyframeProperties{TranslateY: 0, ScaleX: 1, ScaleY: 1}},
		},
	}

	// Shake animation - much more visible
	AnimShake = &Animation{
		Name:           "shake",
		Duration:       600 * time.Millisecond,
		IterationCount: 1,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateX: 0}},
			{Percent: 10, Properties: KeyframeProperties{TranslateX: -25}},
			{Percent: 20, Properties: KeyframeProperties{TranslateX: 25}},
			{Percent: 30, Properties: KeyframeProperties{TranslateX: -25}},
			{Percent: 40, Properties: KeyframeProperties{TranslateX: 25}},
			{Percent: 50, Properties: KeyframeProperties{TranslateX: -20}},
			{Percent: 60, Properties: KeyframeProperties{TranslateX: 20}},
			{Percent: 70, Properties: KeyframeProperties{TranslateX: -15}},
			{Percent: 80, Properties: KeyframeProperties{TranslateX: 15}},
			{Percent: 90, Properties: KeyframeProperties{TranslateX: -5}},
			{Percent: 100, Properties: KeyframeProperties{TranslateX: 0}},
		},
	}

	// Slide animations
	AnimSlideInLeft = &Animation{
		Name:       "slideInLeft",
		Duration:   400 * time.Millisecond,
		TimingFunc: EaseOutCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateX: -100, Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{TranslateX: 0, Opacity: 1}},
		},
	}

	AnimSlideInRight = &Animation{
		Name:       "slideInRight",
		Duration:   400 * time.Millisecond,
		TimingFunc: EaseOutCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateX: 100, Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{TranslateX: 0, Opacity: 1}},
		},
	}

	AnimSlideInUp = &Animation{
		Name:       "slideInUp",
		Duration:   400 * time.Millisecond,
		TimingFunc: EaseOutCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateY: 100, Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{TranslateY: 0, Opacity: 1}},
		},
	}

	AnimSlideInDown = &Animation{
		Name:       "slideInDown",
		Duration:   400 * time.Millisecond,
		TimingFunc: EaseOutCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateY: -100, Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{TranslateY: 0, Opacity: 1}},
		},
	}

	// Zoom animations
	AnimZoomIn = &Animation{
		Name:       "zoomIn",
		Duration:   300 * time.Millisecond,
		TimingFunc: EaseOutCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{ScaleX: 0.3, ScaleY: 0.3, Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1, Opacity: 1}},
		},
	}

	AnimZoomOut = &Animation{
		Name:       "zoomOut",
		Duration:   300 * time.Millisecond,
		TimingFunc: EaseInCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1, Opacity: 1}},
			{Percent: 100, Properties: KeyframeProperties{ScaleX: 0.3, ScaleY: 0.3, Opacity: 0}},
		},
	}

	// Rotate animations
	AnimRotateIn = &Animation{
		Name:       "rotateIn",
		Duration:   500 * time.Millisecond,
		TimingFunc: EaseOutCubic,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{Rotate: -200, Opacity: 0}},
			{Percent: 100, Properties: KeyframeProperties{Rotate: 0, Opacity: 1}},
		},
	}

	// Glow/Shimmer continuous animation
	AnimGlow = &Animation{
		Name:           "glow",
		Duration:       2000 * time.Millisecond,
		IterationCount: -1,
		Direction:      AnimationAlternate,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{BoxShadowBlur: 5, BoxShadowSpread: 0}},
			{Percent: 100, Properties: KeyframeProperties{BoxShadowBlur: 20, BoxShadowSpread: 5}},
		},
	}

	// Heartbeat animation
	AnimHeartbeat = &Animation{
		Name:           "heartbeat",
		Duration:       1300 * time.Millisecond,
		IterationCount: -1,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1}},
			{Percent: 14, Properties: KeyframeProperties{ScaleX: 1.3, ScaleY: 1.3}},
			{Percent: 28, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1}},
			{Percent: 42, Properties: KeyframeProperties{ScaleX: 1.3, ScaleY: 1.3}},
			{Percent: 70, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1}},
			{Percent: 100, Properties: KeyframeProperties{ScaleX: 1, ScaleY: 1}},
		},
	}

	// Wobble animation - more visible swinging
	AnimWobble = &Animation{
		Name:           "wobble",
		Duration:       800 * time.Millisecond,
		IterationCount: 1,
		Keyframes: []Keyframe{
			{Percent: 0, Properties: KeyframeProperties{TranslateX: 0, ScaleX: 1, ScaleY: 1}},
			{Percent: 15, Properties: KeyframeProperties{TranslateX: -40, ScaleX: 1.1, ScaleY: 0.9}},
			{Percent: 30, Properties: KeyframeProperties{TranslateX: 30, ScaleX: 0.95, ScaleY: 1.05}},
			{Percent: 45, Properties: KeyframeProperties{TranslateX: -20, ScaleX: 1.05, ScaleY: 0.95}},
			{Percent: 60, Properties: KeyframeProperties{TranslateX: 15, ScaleX: 0.98, ScaleY: 1.02}},
			{Percent: 75, Properties: KeyframeProperties{TranslateX: -8, ScaleX: 1.01, ScaleY: 0.99}},
			{Percent: 100, Properties: KeyframeProperties{TranslateX: 0, ScaleX: 1, ScaleY: 1}},
		},
	}
)

// ============================================================================
// Animation State Manager
// ============================================================================

// AnimationState tracks the runtime state of an animation on a widget
type AnimationState struct {
	Animation    *Animation
	StartTime    time.Time
	IsPlaying    bool
	IsPaused     bool
	PausedAt     time.Duration
	CurrentProps KeyframeProperties
	Iteration    int
	OnComplete   func()
}

// Start begins playing the animation
func (as *AnimationState) Start() {
	as.StartTime = time.Now()
	as.IsPlaying = true
	as.IsPaused = false
	as.Iteration = 0
}

// Pause pauses the animation
func (as *AnimationState) Pause() {
	if as.IsPlaying && !as.IsPaused {
		as.IsPaused = true
		as.PausedAt = time.Since(as.StartTime)
	}
}

// Resume resumes a paused animation
func (as *AnimationState) Resume() {
	if as.IsPaused {
		as.StartTime = time.Now().Add(-as.PausedAt)
		as.IsPaused = false
	}
}

// Stop stops the animation
func (as *AnimationState) Stop() {
	as.IsPlaying = false
	as.IsPaused = false
}

// Update updates the animation state and returns current properties
func (as *AnimationState) Update() KeyframeProperties {
	if !as.IsPlaying || as.IsPaused || as.Animation == nil {
		return as.CurrentProps
	}

	elapsed := time.Since(as.StartTime) - as.Animation.Delay
	if elapsed < 0 {
		// Still in delay period
		return as.Animation.Keyframes[0].Properties
	}

	duration := as.Animation.Duration
	if duration == 0 {
		duration = 300 * time.Millisecond
	}

	// Calculate progress
	totalElapsed := elapsed
	iterationProgress := float64(elapsed) / float64(duration)

	// Handle iterations
	currentIteration := int(iterationProgress)
	if as.Animation.IterationCount >= 0 && currentIteration >= as.Animation.IterationCount {
		// Animation complete
		as.IsPlaying = false
		if as.OnComplete != nil {
			as.OnComplete()
		}
		// Return final frame based on fill mode
		if as.Animation.FillMode == AnimationFillForwards || as.Animation.FillMode == AnimationFillBoth {
			return as.Animation.Keyframes[len(as.Animation.Keyframes)-1].Properties
		}
		return KeyframeProperties{ScaleX: 1, ScaleY: 1, Opacity: 1}
	}

	// Get progress within current iteration (0-1)
	progress := iterationProgress - float64(currentIteration)

	// Handle direction
	if as.Animation.Direction == AnimationReverse {
		progress = 1 - progress
	} else if as.Animation.Direction == AnimationAlternate {
		if currentIteration%2 == 1 {
			progress = 1 - progress
		}
	} else if as.Animation.Direction == AnimationAlternateReverse {
		if currentIteration%2 == 0 {
			progress = 1 - progress
		}
	}

	// Apply easing
	if as.Animation.TimingFunc != nil {
		progress = as.Animation.TimingFunc(progress)
	}

	// Interpolate between keyframes
	as.CurrentProps = interpolateKeyframes(as.Animation.Keyframes, progress)
	as.Iteration = currentIteration
	_ = totalElapsed // suppress unused warning

	return as.CurrentProps
}

// interpolateKeyframes finds the interpolated properties at a given progress (0-1)
func interpolateKeyframes(keyframes []Keyframe, progress float64) KeyframeProperties {
	if len(keyframes) == 0 {
		return KeyframeProperties{ScaleX: 1, ScaleY: 1, Opacity: 1}
	}
	if len(keyframes) == 1 {
		return keyframes[0].Properties
	}

	// Convert progress (0-1) to percent (0-100)
	percent := progress * 100

	// Find surrounding keyframes
	var prev, next Keyframe
	prev = keyframes[0]
	next = keyframes[len(keyframes)-1]

	for i := 0; i < len(keyframes)-1; i++ {
		if keyframes[i].Percent <= percent && keyframes[i+1].Percent >= percent {
			prev = keyframes[i]
			next = keyframes[i+1]
			break
		}
	}

	// Calculate local progress between the two keyframes
	range_ := next.Percent - prev.Percent
	if range_ == 0 {
		return prev.Properties
	}
	localProgress := (percent - prev.Percent) / range_

	// Interpolate properties
	return KeyframeProperties{
		TranslateX:      lerp(prev.Properties.TranslateX, next.Properties.TranslateX, localProgress),
		TranslateY:      lerp(prev.Properties.TranslateY, next.Properties.TranslateY, localProgress),
		ScaleX:          lerpWithDefault(prev.Properties.ScaleX, next.Properties.ScaleX, localProgress, 1),
		ScaleY:          lerpWithDefault(prev.Properties.ScaleY, next.Properties.ScaleY, localProgress, 1),
		Rotate:          lerp(prev.Properties.Rotate, next.Properties.Rotate, localProgress),
		SkewX:           lerp(prev.Properties.SkewX, next.Properties.SkewX, localProgress),
		SkewY:           lerp(prev.Properties.SkewY, next.Properties.SkewY, localProgress),
		Opacity:         lerpWithDefault(prev.Properties.Opacity, next.Properties.Opacity, localProgress, 1),
		BoxShadowBlur:   lerp(prev.Properties.BoxShadowBlur, next.Properties.BoxShadowBlur, localProgress),
		BoxShadowSpread: lerp(prev.Properties.BoxShadowSpread, next.Properties.BoxShadowSpread, localProgress),
		Width:           lerp(prev.Properties.Width, next.Properties.Width, localProgress),
		Height:          lerp(prev.Properties.Height, next.Properties.Height, localProgress),
		OffsetX:         lerp(prev.Properties.OffsetX, next.Properties.OffsetX, localProgress),
		OffsetY:         lerp(prev.Properties.OffsetY, next.Properties.OffsetY, localProgress),
	}
}

// lerp linearly interpolates between two values
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// lerpWithDefault lerps but uses default if both values are 0
func lerpWithDefault(a, b, t, def float64) float64 {
	if a == 0 && b == 0 {
		return def
	}
	if a == 0 {
		a = def
	}
	if b == 0 {
		b = def
	}
	return a + (b-a)*t
}

// ============================================================================
// Transition System
// ============================================================================

// TransitionManager handles smooth property transitions
type TransitionManager struct {
	transitions map[string]*PropertyTransition
}

// PropertyTransition represents a single property transition
type PropertyTransition struct {
	Property   string
	StartValue float64
	EndValue   float64
	StartTime  time.Time
	Duration   time.Duration
	Delay      time.Duration
	Easing     EasingFunc
	IsActive   bool
}

// NewTransitionManager creates a new transition manager
func NewTransitionManager() *TransitionManager {
	return &TransitionManager{
		transitions: make(map[string]*PropertyTransition),
	}
}

// StartTransition begins a property transition
func (tm *TransitionManager) StartTransition(property string, from, to float64, duration time.Duration, easing EasingFunc) {
	tm.transitions[property] = &PropertyTransition{
		Property:   property,
		StartValue: from,
		EndValue:   to,
		StartTime:  time.Now(),
		Duration:   duration,
		Easing:     easing,
		IsActive:   true,
	}
}

// GetValue returns the current interpolated value for a property
func (tm *TransitionManager) GetValue(property string, defaultValue float64) float64 {
	t, exists := tm.transitions[property]
	if !exists || !t.IsActive {
		return defaultValue
	}

	elapsed := time.Since(t.StartTime) - t.Delay
	if elapsed < 0 {
		return t.StartValue
	}

	progress := float64(elapsed) / float64(t.Duration)
	if progress >= 1 {
		t.IsActive = false
		return t.EndValue
	}

	if t.Easing != nil {
		progress = t.Easing(progress)
	}

	return lerp(t.StartValue, t.EndValue, progress)
}

// IsTransitioning returns true if any transition is active
func (tm *TransitionManager) IsTransitioning() bool {
	for _, t := range tm.transitions {
		if t.IsActive {
			return true
		}
	}
	return false
}

// ============================================================================
// Ripple Effect (Material Design-like)
// ============================================================================

// RippleEffect represents a ripple animation at a click point
type RippleEffect struct {
	X, Y      float64
	MaxRadius float64
	StartTime time.Time
	Duration  time.Duration
	Color     color.Color
	IsActive  bool
}

// NewRippleEffect creates a new ripple effect at the given position
func NewRippleEffect(x, y, maxRadius float64, clr color.Color) *RippleEffect {
	return &RippleEffect{
		X:         x,
		Y:         y,
		MaxRadius: maxRadius,
		StartTime: time.Now(),
		Duration:  600 * time.Millisecond,
		Color:     clr,
		IsActive:  true,
	}
}

// Update returns current radius and alpha for the ripple
func (r *RippleEffect) Update() (radius float64, alpha float64) {
	if !r.IsActive {
		return 0, 0
	}

	elapsed := time.Since(r.StartTime)
	progress := float64(elapsed) / float64(r.Duration)

	if progress >= 1 {
		r.IsActive = false
		return 0, 0
	}

	// Ease out for smooth expansion
	easedProgress := EaseOutCubic(progress)

	radius = r.MaxRadius * easedProgress
	alpha = 1 - progress // Fade out as it expands

	return radius, alpha
}

// ============================================================================
// Shimmer Effect (Loading skeleton animation)
// ============================================================================

// ShimmerEffect creates a loading skeleton shimmer animation
type ShimmerEffect struct {
	StartTime time.Time
	Duration  time.Duration
	IsActive  bool
}

// NewShimmerEffect creates a new shimmer effect
func NewShimmerEffect() *ShimmerEffect {
	return &ShimmerEffect{
		StartTime: time.Now(),
		Duration:  1500 * time.Millisecond,
		IsActive:  true,
	}
}

// GetOffset returns the current shimmer offset (0-1 position of the highlight)
func (s *ShimmerEffect) GetOffset() float64 {
	if !s.IsActive {
		return 0
	}

	elapsed := time.Since(s.StartTime)
	progress := float64(elapsed) / float64(s.Duration)

	// Loop the progress
	progress = math.Mod(progress, 1.0)

	return progress
}

// ============================================================================
// Animation Registry
// ============================================================================

var animationRegistry = map[string]*Animation{
	"fadeIn":       AnimFadeIn,
	"fadeOut":      AnimFadeOut,
	"pulse":        AnimPulse,
	"bounce":       AnimBounce,
	"shake":        AnimShake,
	"slideInLeft":  AnimSlideInLeft,
	"slideInRight": AnimSlideInRight,
	"slideInUp":    AnimSlideInUp,
	"slideInDown":  AnimSlideInDown,
	"zoomIn":       AnimZoomIn,
	"zoomOut":      AnimZoomOut,
	"rotateIn":     AnimRotateIn,
	"glow":         AnimGlow,
	"heartbeat":    AnimHeartbeat,
	"wobble":       AnimWobble,
}

// GetAnimation returns a pre-defined animation by name
func GetAnimation(name string) *Animation {
	if anim, exists := animationRegistry[name]; exists {
		// Return a copy so each usage has independent state
		copied := *anim
		return &copied
	}
	return nil
}

// RegisterAnimation registers a custom animation
func RegisterAnimation(name string, anim *Animation) {
	animationRegistry[name] = anim
}

// ListAnimations returns all available animation names
func ListAnimations() []string {
	names := make([]string, 0, len(animationRegistry))
	for name := range animationRegistry {
		names = append(names, name)
	}
	return names
}
