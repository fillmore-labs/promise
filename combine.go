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

// AnyFuture matches a [Future] of any type.
type AnyFuture interface {
	reflect() reflect.Value
}

// AwaitAll returns a function that yields the results of all futures.
// If the context is canceled, it returns an error for the remaining futures.
func AwaitAll[R any](ctx context.Context, futures ...Future[R]) func(yield func(int, result.Result[R]) bool) {
	i := newIterator(ctx, convertValue[R], futures)

	return i.yieldTo
}

// AwaitAllAny returns a function that yields the results of all futures.
// If the context is canceled, it returns an error for the remaining futures.
func AwaitAllAny(ctx context.Context, futures ...AnyFuture) func(yield func(int, result.Result[any]) bool) {
	i := newIterator(ctx, convertValueAny, futures)

	return i.yieldTo
}

// AwaitAllResults waits for all futures to complete and returns the results.
// If the context is canceled, it returns early with errors for the remaining futures.
func AwaitAllResults[R any](ctx context.Context, futures ...Future[R]) []result.Result[R] {
	return awaitAllResults(len(futures), AwaitAll(ctx, futures...))
}

// AwaitAllResultsAny waits for all futures to complete and returns the results.
// If the context is canceled, it returns early with errors for the remaining futures.
func AwaitAllResultsAny(ctx context.Context, futures ...AnyFuture) []result.Result[any] {
	return awaitAllResults(len(futures), AwaitAllAny(ctx, futures...))
}

func awaitAllResults[R any](n int, iter func(yield func(int, result.Result[R]) bool)) []result.Result[R] {
	results := make([]result.Result[R], n)

	iter(func(i int, r result.Result[R]) bool {
		results[i] = r

		return true
	})

	return results
}

// AwaitAllValues returns the values of completed futures.
// If any future fails or the context is canceled, it returns early with an error.
func AwaitAllValues[R any](ctx context.Context, futures ...Future[R]) ([]R, error) {
	return awaitAllValues(len(futures), AwaitAll(ctx, futures...))
}

// AwaitAllValuesAny returns the values of completed futures.
// If any future fails or the context is canceled, it returns early with an error.
func AwaitAllValuesAny(ctx context.Context, futures ...AnyFuture) ([]any, error) {
	return awaitAllValues(len(futures), AwaitAllAny(ctx, futures...))
}

func awaitAllValues[R any](n int, iter func(yield func(int, result.Result[R]) bool)) ([]R, error) {
	results := make([]R, n)
	var yieldErr error

	iter(func(i int, r result.Result[R]) bool {
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
func AwaitFirst[R any](ctx context.Context, futures ...Future[R]) (R, error) {
	return awaitFirst(AwaitAll(ctx, futures...))
}

// AwaitFirstAny returns the result of the first completed future.
// If the context is canceled, it returns early with an error.
func AwaitFirstAny(ctx context.Context, futures ...AnyFuture) (any, error) {
	return awaitFirst(AwaitAllAny(ctx, futures...))
}

func awaitFirst[R any](iter func(yield func(int, result.Result[R]) bool)) (R, error) {
	var v result.Result[R]

	iter(func(_ int, r result.Result[R]) bool {
		v = r

		return false
	})

	if v == nil {
		return *new(R), ErrNoResult
	}

	return v.V()
}
