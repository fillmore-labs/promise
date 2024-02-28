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

package promise

import (
	"context"
	"fmt"
	"reflect"

	"fillmore-labs.com/promise/result"
)

// Future represents an asynchronous operation that will complete sometime in the future.
//
// It is a read-only channel that can be used with [Future.Await] to retrieve the final result of a
// [Promise].
type Future[R any] <-chan result.Result[R]

// NewAsync runs fn asynchronously, immediately returning a [Future] that can be used to retrieve
// the eventual result. This allows separating evaluating the result from computation.
func NewAsync[R any](fn func() (R, error)) Future[R] {
	p, f := New[R]()
	go p.Do(fn)

	return f
}

// Await returns the final result of the associated [Promise].
// It can only be called once and blocks until a result is received or the context is canceled.
// If you need to read multiple times from a [Future] wrap it with [Future.Memoize].
func (f Future[R]) Await(ctx context.Context) (R, error) {
	select {
	case r, ok := <-f:
		if !ok || r == nil {
			return *new(R), ErrNoResult
		}

		return r.V()

	case <-ctx.Done():
		return *new(R), fmt.Errorf("channel await: %w", context.Cause(ctx))
	}
}

// Try returns the result when ready, [ErrNotReady] otherwise.
func (f Future[R]) Try() (R, error) {
	select {
	case r, ok := <-f:
		if !ok || r == nil {
			return *new(R), ErrNoResult
		}

		return r.V()

	default:
		return *new(R), ErrNotReady
	}
}

func (f Future[_]) reflect() reflect.Value {
	return reflect.ValueOf(f)
}
