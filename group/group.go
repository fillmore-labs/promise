// Copyright 2023-2024 Oliver Eikemeier. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package group

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"fillmore-labs.com/promise"
	"golang.org/x/sync/semaphore"
)

// A Group represents a set of collaborating goroutines within a common overarching task.
//
// Should this group have a limit of active goroutines or cancel on the first error it needs to be created with [New].
type Group struct {
	sema   *semaphore.Weighted
	err    atomic.Pointer[error]
	cancel func(error)
	panic  atomic.Value
	wg     sync.WaitGroup
}

// New creates a new [Group] with the given options.
func New(opts ...Option) *Group {
	var option options
	for _, opt := range opts {
		opt.apply(&option)
	}

	var sema *semaphore.Weighted
	if option.limit > 0 {
		sema = semaphore.NewWeighted(int64(option.limit))
	}

	return &Group{sema: sema, cancel: option.cancel}
}

// options defines configurable parameters for the group.
type options struct {
	cancel context.CancelCauseFunc
	limit  int
}

// Option defines configurations for [New].
type Option interface {
	apply(opts *options)
}

// WithLimit is an [Option] to configure the limit of active goroutines.
func WithLimit(limit int) Option {
	if limit < 1 {
		panic("limit must be at least 1")
	}

	return limitOption{limit: limit}
}

type limitOption struct {
	limit int
}

func (o limitOption) apply(opts *options) {
	opts.limit = o.limit
}

// WithCancel is an [Option] to cancel a context on the first error.
//
// cancel is a function retrieved from [context.WithCancelCause].
func WithCancel(cancel context.CancelCauseFunc) Option {
	return cancelOption{cancel: cancel}
}

type cancelOption struct {
	cancel context.CancelCauseFunc
}

func (o cancelOption) apply(opts *options) {
	opts.cancel = o.cancel
}

// Wait blocks until all goroutines spawned from [Group.Go] and [DoAsync] are finished and
// returns the first non-nil error from them.
func (g *Group) Wait() error {
	g.wg.Wait()

	if p := g.panic.Load(); p != nil {
		panic(p)
	}

	if err := g.err.Load(); err != nil {
		return *err
	}

	return nil
}

type ExecutionError struct {
	Value any
}

func (e ExecutionError) Error() string {
	return fmt.Sprintf("execution error: %v", e.Value)
}

// DoAsync calls the given function in a new goroutine.
//
// If there is a limit on active goroutines within the group, it blocks until it can be spawned without surpassing the
// limit. If the passed context is canceled, a failed future is returned.
//
// If the group was created with [Option] [WithCancel], the first call that returns a non-nil error will cancel the
// group's context. This error will subsequently be returned by [Group.Wait].
func DoAsync[R any](ctx context.Context, g *Group, fn func() (R, error)) promise.Future[R] {
	p, f := promise.New[R]()

	if err := g.acquire(ctx); err != nil {
		p.Reject(err)

		return f
	}

	go p.Do(func() (R, error) {
		defer g.release()

		value, err := func() (value R, err error) {
			defer g.recover(&err)

			return fn()
		}()
		if err != nil {
			g.setError(err)
		}

		return value, err
	})

	return f
}

// Go calls the given function in a new goroutine.
//
// It is here for compatibility reasons and calls [Group.GoCtx] with [context.TODO].
// You should consider calling that function instead.
func (g *Group) Go(fn func() error) {
	g.GoCtx(context.TODO(), fn)
}

// GoCtx calls the given function in a new goroutine.
//
// If there is a limit on active goroutines within the group, it blocks until it can be spawned without surpassing the
// limit. If the passed context is canceled, an error is returned.
//
// If the group was created with [Option] [WithCancel], the first call that returns a non-nil error will cancel the
// group's context. This error will subsequently be returned by [Group.Wait].
func (g *Group) GoCtx(ctx context.Context, fn func() error) {
	if err := g.acquire(ctx); err != nil {
		g.setError(err)

		return
	}

	go func() {
		defer g.release()

		g.Do(fn)
	}()
}

// Do calls the given function synchronously.
//
// If the group was created with [Option] [WithCancel], the first call that returns a non-nil error will cancel the
// group's context. This error will subsequently be returned by [Group.Wait].
func (g *Group) Do(fn func() error) {
	err := func() (err error) {
		defer g.recover(&err)

		return fn()
	}()
	if err != nil {
		g.setError(err)
	}
}

func (g *Group) acquire(ctx context.Context) error {
	if g.sema != nil {
		if err := g.sema.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("can not schedule goroutine: %w", err)
		}
	}
	g.wg.Add(1)

	return nil
}

func (g *Group) release() {
	if g.sema != nil {
		g.sema.Release(1)
	}
	g.wg.Done()
}

func (g *Group) setError(err error) {
	if g.err.CompareAndSwap(nil, &err) {
		if g.cancel != nil {
			g.cancel(err)
		}
	}
}

func (g *Group) recover(err *error) {
	if r := recover(); r != nil {
		_ = g.panic.CompareAndSwap(nil, r)
		*err = ExecutionError{Value: r}
	}
}
