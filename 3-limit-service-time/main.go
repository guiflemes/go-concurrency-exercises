//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"context"
	"sync/atomic"
	"time"
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  atomic.Uint64
}

// Advanced Level: 10s max per user (accumulated)
func HandleRequest(process func(), u *User) bool {

	done := make(chan struct{})
	timeout := make(chan struct{})

	go func() {
		defer close(done)
		start := time.Now()
		process()
		u.TimeUsed.Add(uint64(time.Since(start).Seconds()))

		if u.TimeUsed.Load() > 10 {
			timeout <- struct{}{}
		}

		done <- struct{}{}
	}()

	if u.IsPremium {
		return true
	}

	select {

	case <-timeout:
		return false

	case <-done:
		return true
	}

}

// Beginner Level: 10s max per request
func HandleRequestX(process func(), u *User) bool {

	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		defer close(done)
		process()
		done <- struct{}{}
	}()

	if u.IsPremium {
		return true
	}

	select {
	case <-ctx.Done():
		return false
	case <-done:

		return true
	}

}

func main() {
	RunMockServer()
}
