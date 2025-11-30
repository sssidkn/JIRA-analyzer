package ratelimiter

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	maxDelay := 1 * time.Second
	rl := NewRateLimiter(baseDelay, maxDelay)

	if rl.BaseDelay != baseDelay {
		t.Errorf("Expected BaseDelay %v, got %v", baseDelay, rl.BaseDelay)
	}
	if rl.MaxDelay != maxDelay {
		t.Errorf("Expected MaxDelay %v, got %v", maxDelay, rl.MaxDelay)
	}
	if rl.CurrentDelay != baseDelay {
		t.Errorf("Expected CurrentDelay %v, got %v", baseDelay, rl.CurrentDelay)
	}
	if rl.paused != false {
		t.Errorf("Expected paused false, got %v", rl.paused)
	}
	if rl.notifyPause == nil {
		t.Error("Expected notifyPause channel to be initialized")
	}
}

func TestShouldPause_WhenNotPaused(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	paused, remaining := rl.ShouldPause()

	if paused != false {
		t.Errorf("Expected paused false, got %v", paused)
	}
	if remaining != 0 {
		t.Errorf("Expected remaining 0, got %v", remaining)
	}
}

func TestShouldPause_WhenPausedAndTimeNotExpired(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	// Manually set paused state
	rl.m.Lock()
	rl.paused = true
	rl.pauseUntil = time.Now().Add(50 * time.Millisecond)
	rl.m.Unlock()

	paused, remaining := rl.ShouldPause()

	if paused != true {
		t.Errorf("Expected paused true, got %v", paused)
	}
	if remaining <= 0 {
		t.Errorf("Expected remaining > 0, got %v", remaining)
	}
}

func TestShouldPause_WhenPausedAndTimeExpired(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	// Manually set paused state with expired time
	rl.m.Lock()
	rl.paused = true
	rl.pauseUntil = time.Now().Add(-50 * time.Millisecond) // past time
	rl.m.Unlock()

	paused, remaining := rl.ShouldPause()

	if paused != false {
		t.Errorf("Expected paused false, got %v", paused)
	}
	if remaining != 0 {
		t.Errorf("Expected remaining 0, got %v", remaining)
	}

	// Verify internal state was updated
	if rl.paused != false {
		t.Errorf("Expected internal paused state to be false after expiration")
	}
}

func TestPause(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)
	initialDelay := rl.CurrentDelay

	// Capture the notify channel before pause
	notifyChan := rl.NotifyPause()

	rl.Pause()

	// Verify paused state
	if !rl.paused {
		t.Error("Expected paused true after Pause()")
	}

	// Verify pauseUntil is set correctly
	expectedPauseUntil := time.Now().Add(initialDelay)
	if rl.pauseUntil.Before(expectedPauseUntil.Add(-10*time.Millisecond)) ||
		rl.pauseUntil.After(expectedPauseUntil.Add(10*time.Millisecond)) {
		t.Errorf("pauseUntil not set correctly, expected around %v, got %v",
			expectedPauseUntil, rl.pauseUntil)
	}

	// Verify delay was doubled
	expectedDelay := initialDelay * 2
	if rl.CurrentDelay != expectedDelay {
		t.Errorf("Expected CurrentDelay %v, got %v", expectedDelay, rl.CurrentDelay)
	}

	// Verify notify channel was closed and new one created
	select {
	case <-notifyChan:
		// Channel was closed, which is expected
	default:
		t.Error("Expected original notifyPause channel to be closed")
	}

	if rl.notifyPause == notifyChan {
		t.Error("Expected new notifyPause channel to be created")
	}
}

func TestPause_WithMaxDelay(t *testing.T) {
	baseDelay := 500 * time.Millisecond
	maxDelay := 1 * time.Second
	rl := NewRateLimiter(baseDelay, maxDelay)

	// First pause should double to 1s (exactly max)
	rl.Pause()
	if rl.CurrentDelay != maxDelay {
		t.Errorf("Expected CurrentDelay %v, got %v", maxDelay, rl.CurrentDelay)
	}

	// Second pause should stay at max (not exceed it)
	rl.Pause()
	if rl.CurrentDelay != maxDelay {
		t.Errorf("Expected CurrentDelay to remain at %v, got %v", maxDelay, rl.CurrentDelay)
	}
}

func TestReset(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	// Set up some state
	rl.Pause()
	rl.Pause() // Increase delay

	rl.Reset()

	// Verify reset state
	if rl.paused != false {
		t.Error("Expected paused false after Reset")
	}
	if !rl.pauseUntil.IsZero() {
		t.Error("Expected pauseUntil to be zero time after Reset")
	}
	if rl.CurrentDelay != rl.BaseDelay {
		t.Errorf("Expected CurrentDelay %v, got %v", rl.BaseDelay, rl.CurrentDelay)
	}
}

func TestNotifyPause(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	channel1 := rl.NotifyPause()
	channel2 := rl.NotifyPause()

	if channel1 != channel2 {
		t.Error("Expected NotifyPause to return the same channel instance")
	}
}

func TestPaused(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	// Initially not paused
	if rl.Paused() != false {
		t.Error("Expected Paused() to return false initially")
	}

	// After pause
	rl.Pause()
	if rl.Paused() != true {
		t.Error("Expected Paused() to return true after Pause()")
	}

	// After reset
	rl.Reset()
	if rl.Paused() != false {
		t.Error("Expected Paused() to return false after Reset()")
	}
}

func TestConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1*time.Second)

	done := make(chan bool)

	// Start multiple goroutines accessing the rate limiter
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				rl.ShouldPause()
				rl.Paused()
				if j%10 == 0 {
					rl.Pause()
				}
				if j%20 == 0 {
					rl.Reset()
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Final state should be consistent
	if rl.paused && time.Now().After(rl.pauseUntil) {
		t.Error("Rate limiter in inconsistent state after concurrent access")
	}
}
