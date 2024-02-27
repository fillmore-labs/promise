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

package promise_test

import (
	"context"
	"testing"

	"fillmore-labs.com/promise"
	"fillmore-labs.com/promise/result"
	"github.com/stretchr/testify/assert"
)

const iterations = 3

func makePromisesAndFutures[R any]() ([]promise.Promise[R], promise.List[R]) {
	var promises [iterations]promise.Promise[R]
	var futures [iterations]promise.Future[R]

	for i := 0; i < iterations; i++ {
		promises[i], futures[i] = promise.New[R]()
	}

	return promises[:], futures[:]
}

func TestWaitAll(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()

	promises[0].Resolve(1)
	promises[1].Reject(errTest)
	close(promises[2])

	// when
	ctx := context.Background()
	results := futures.AwaitAll(ctx)

	// then
	assert.Len(t, results, len(futures))
	v0, err0 := results[0].V()
	_, err1 := results[1].V()
	_, err2 := results[2].V()

	if assert.NoError(t, err0) {
		assert.Equal(t, 1, v0)
	}
	assert.ErrorIs(t, err1, errTest)
	assert.ErrorIs(t, err2, promise.ErrNoResult)
}

func TestAllValues(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	for i := 0; i < iterations; i++ {
		promises[i].Resolve(i + 1)
	}

	// when
	ctx := context.Background()
	results, err := futures.AwaitAllValues(ctx)

	// then
	if assert.NoError(t, err) {
		assert.Len(t, results, iterations)
		for i := 0; i < iterations; i++ {
			assert.Equal(t, i+1, results[i])
		}
	}
}

func TestAllValuesError(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	promises[1].Reject(errTest)

	// when
	ctx := context.Background()
	_, err := futures.AwaitAllValues(ctx)

	// then
	assert.ErrorIs(t, err, errTest)
}

func TestFirst(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	promises[1].Resolve(2)

	// when
	ctx := context.Background()
	result, err := futures.AwaitFirst(ctx)

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 2, result)
	}
}

func TestCombineCancellation(t *testing.T) {
	t.Parallel()

	subTests := []struct {
		name    string
		combine func(promise.List[int], context.Context) error
	}{
		{name: "First", combine: func(futures promise.List[int], ctx context.Context) error {
			_, err := futures.AwaitFirst(ctx)

			return err
		}},
		{name: "All", combine: func(futures promise.List[int], ctx context.Context) error {
			r := futures.AwaitAll(ctx)

			return r[0].Err()
		}},
		{name: "AllValues", combine: func(futures promise.List[int], ctx context.Context) error {
			_, err := futures.AwaitAllValues(ctx)

			return err
		}},
	}

	for _, tc := range subTests {
		combine := tc.combine
		test := func(t *testing.T) {
			t.Helper()
			t.Parallel()

			// given
			_, futures := makePromisesAndFutures[int]()

			// when
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := combine(futures, ctx)

			// then
			assert.ErrorIs(t, err, context.Canceled)
		}

		_ = t.Run(tc.name, test)
	}
}

func TestCombineMemoized(t *testing.T) {
	t.Parallel()

	subTests := []struct {
		name    string
		combine func(promise.List[int], context.Context) (any, error)
		expect  func(t *testing.T, actual any)
	}{
		{name: "First", combine: func(futures promise.List[int], ctx context.Context) (any, error) {
			return futures.AwaitFirst(ctx)
		}, expect: func(t *testing.T, actual any) { t.Helper(); assert.Equal(t, 3, actual) }},
		{name: "All", combine: func(futures promise.List[int], ctx context.Context) (any, error) {
			return futures.AwaitAll(ctx), nil
		}, expect: func(t *testing.T, actual any) {
			t.Helper()
			vv, ok := actual.([]result.Result[int])
			if !ok {
				assert.Fail(t, "Unexpected result type")

				return
			}

			for _, v := range vv {
				value, err := v.V()
				if assert.NoError(t, err) {
					assert.Equal(t, 3, value)
				}
			}
		}},
		{name: "AllValues", combine: func(futures promise.List[int], ctx context.Context) (any, error) {
			return futures.AwaitAllValues(ctx)
		}, expect: func(t *testing.T, actual any) { t.Helper(); assert.Equal(t, []int{3, 3, 3}, actual) }},
	}

	for _, tc := range subTests {
		combine := tc.combine
		expect := tc.expect
		_ = t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			// given
			promises, futures := makePromisesAndFutures[int]()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for _, promise := range promises {
				promise.Resolve(3)
			}

			// when
			result, err := combine(futures, ctx)

			// then
			if assert.NoError(t, err) {
				expect(t, result)
			}
		})
	}
}

func TestAwaitAllEmpty(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var futures promise.List[int]

	// when
	results := futures.AwaitAll(ctx)

	assert.Empty(t, results)
}

func TestAwaitAllValuesEmpty(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var futures promise.List[int]

	// when
	results, err := futures.AwaitAllValues(ctx)

	if assert.NoError(t, err) {
		assert.Empty(t, results)
	}
}

func TestAwaitFirstEmpty(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var futures promise.List[int]

	// when
	_, err := futures.AwaitFirst(ctx)

	assert.ErrorIs(t, err, promise.ErrNoResult)
}
