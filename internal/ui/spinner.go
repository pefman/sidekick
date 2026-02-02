package ui

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	frames  []string
	message string
	mu      sync.Mutex
	active  bool
	done    chan bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		done:    make(chan bool),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				frame := s.frames[i%len(s.frames)]
				msg := s.message
				s.mu.Unlock()

				fmt.Printf("\r\033[K\033[38;5;208m%s\033[0m %s", frame, msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	s.done <- true
	fmt.Print("\r\033[K") // Clear line
}
