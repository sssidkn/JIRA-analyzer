package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	m            sync.Mutex
	paused       bool
	pauseUntil   time.Time
	CurrentDelay time.Duration
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	notifyPause  chan struct{}
}

func (rl *RateLimiter) NotifyPause() chan struct{} {
	return rl.notifyPause
}

func (rl *RateLimiter) Paused() bool {
	return rl.paused
}

func NewRateLimiter(baseDelay, maxDelay time.Duration) *RateLimiter {
	return &RateLimiter{
		m:            sync.Mutex{},
		paused:       false,
		BaseDelay:    baseDelay,
		MaxDelay:     maxDelay,
		CurrentDelay: baseDelay,
		notifyPause:  make(chan struct{}),
	}
}

func (rl *RateLimiter) ShouldPause() (bool, time.Duration) {
	rl.m.Lock()
	defer rl.m.Unlock()

	if !rl.paused {
		return false, 0
	}

	if time.Now().Before(rl.pauseUntil) {
		return true, time.Until(rl.pauseUntil)
	}

	rl.paused = false
	return false, 0
}

func (rl *RateLimiter) Pause() {
	rl.m.Lock()
	defer rl.m.Unlock()

	rl.paused = true
	rl.pauseUntil = time.Now().Add(rl.CurrentDelay)

	nextDelay := rl.CurrentDelay * 2
	rl.CurrentDelay = min(nextDelay, rl.MaxDelay)

	close(rl.notifyPause)
	rl.notifyPause = make(chan struct{})
}

func (rl *RateLimiter) Reset() {
	rl.m.Lock()
	defer rl.m.Unlock()

	rl.paused = false
	rl.pauseUntil = time.Time{}
	rl.CurrentDelay = rl.BaseDelay
}
