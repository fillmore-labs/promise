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
	"runtime/trace"

	"fillmore-labs.com/promise/result"
)

// List is a list of [Future], representing results of asynchronous tasks.
type List[R any] []Future[R]

// All returns a function that yields the results of all futures.
// If the context is canceled, it returns an error for the remaining futures.
func (l List[R]) All(ctx context.Context) func(yield func(int, result.Result[R]) bool) {
	defer trace.StartRegion(ctx, "asyncSeq").End()
	s := newIterator(ctx, l)

	return s.yieldTo
}

// AwaitAll waits for all futures to complete and returns the results.
// If the context is canceled, it returns early with errors for the remaining futures.
func (l List[R]) AwaitAll(ctx context.Context) []result.Result[R] {
	results := make([]result.Result[R], len(l))
	l.All(ctx)(func(i int, r result.Result[R]) bool {
		results[i] = r

		return true
	})

	return results
}

// AwaitAllValues returns the values of completed futures.
// If any future fails or the context is canceled, it returns early with an error.
func (l List[R]) AwaitAllValues(ctx context.Context) ([]R, error) {
	results := make([]R, len(l))
	var yieldErr error
	l.All(ctx)(func(i int, r result.Result[R]) bool {
		if r.Err() != nil {
			yieldErr = fmt.Errorf("list AwaitAllValues result %d: %w", i, r.Err())

			return false
		}
		results[i] = r.Value()

		return true
	})

	return results, yieldErr
}

// AwaitFirst returns the result of the first completed future.
// If the context is canceled, it returns early with an error.
func (l List[R]) AwaitFirst(ctx context.Context) (R, error) {
	var v result.Result[R]
	l.All(ctx)(func(_ int, r result.Result[R]) bool {
		v = r

		return false
	})
	if v == nil {
		return *new(R), ErrNoResult
	}

	return v.V()
}
