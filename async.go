package main

import (
	"sync/atomic"
	"time"
)

// goLimit allows for simplified concurrent task distribution and await.
//
// Provided input `maxJobs` sets the ceiling on the number of concurrent tasks.
//
// The first closure returned (`spawn`) allows submission of any closure, which
// will spawn ('go') when the number of concurrent workers falls below
// `maxJobs`.
//
// The second closure (`wait`) blocks ('wait's) until all active jobs have
// completed.
func goLimit(maxJobs int32) (spawn func(fn func()), wait func()) {
	jobs := new(atomic.Int32)

	return func(fn func()) {
			for jobs.Load() >= maxJobs {
				<-time.After(1 * time.Millisecond)
			}
			jobs.Add(1)
			go func() {
				defer jobs.Add(-1)
				fn()
			}()
		}, func() {
			for jobs.Load() > 0 {
				<-time.After(1 * time.Millisecond)
			}
		}
}
