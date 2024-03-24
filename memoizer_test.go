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
	"sync"
	"testing"
	"time"

	"fillmore-labs.com/promise"
	"github.com/stretchr/testify/assert"
)

func TestMemoizerTry(t *testing.T) {
	t.Parallel()

	// given
	p, f := promise.New[int]()
	m := f.Memoize()

	// when
	_, err1 := m.Try()

	p.Resolve(1)
	value2, err2 := m.Try()
	value3, err3 := m.Try()

	// then
	assert.ErrorIs(t, err1, promise.ErrNotReady)
	if assert.NoError(t, err2) {
		assert.Equal(t, 1, value2)
	}
	if assert.NoError(t, err3) {
		assert.Equal(t, 1, value3)
	}
}

func TestMemoizerTryConcurrent(t *testing.T) {
	t.Parallel()

	// given
	p, f := promise.New[int]()
	m := f.Memoize()

	// when
	ctx := context.Background()

	f1 := promise.NewAsync(func() (int, error) { return m.Await(ctx) })
	time.Sleep(time.Millisecond)

	_, err1 := m.Try()

	p.Resolve(1)
	value2, err2 := f1.Await(ctx)

	// then
	assert.ErrorIs(t, err1, promise.ErrNotReady)
	if assert.NoError(t, err2) {
		assert.Equal(t, 1, value2)
	}
}

func TestMemoizerMany(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 1_000
	p, f := promise.New[int]()
	m := f.Memoize()

	// when
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := make([]promise.Result[int], iterations)
	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = promise.NewResult(m.Await(ctx))
		}(i)
	}

	p.Resolve(1)
	wg.Wait()

	// then
	for i := 0; i < iterations; i++ {
		if assert.NoError(t, results[i].Err) {
			assert.Equal(t, 1, results[i].Value)
		}
	}
}

func TestMemoizerCancel(t *testing.T) {
	t.Parallel()

	// given
	_, f := promise.New[int]()
	m := f.Memoize()

	// when
	ctx := context.Background()

	ctx1, cancel1 := context.WithCancel(ctx)
	ctx2, cancel2 := context.WithCancel(ctx)

	fn1 := func() (int, error) { return m.Await(ctx1) }
	fn2 := func() (int, error) { return m.Await(ctx2) }

	f1 := promise.NewAsync(fn1)
	time.Sleep(time.Millisecond)
	f2 := promise.NewAsync(fn2)

	cancel2()
	_, err1 := f2.Await(ctx)
	_, err2 := f1.Try()

	cancel1()
	_, err3 := f1.Await(ctx)

	// then
	assert.ErrorIs(t, err1, context.Canceled)
	assert.ErrorIs(t, err2, promise.ErrNotReady)
	assert.ErrorIs(t, err3, context.Canceled)
}

func TestMemoizerClosed(t *testing.T) {
	t.Parallel()

	// given
	p, f := promise.New[int]()
	close(p)

	m := f.Memoize()

	ctx := context.Background()

	// when
	_, err := m.Await(ctx)

	// then
	assert.ErrorIs(t, err, promise.ErrNoResult)
}

func TestMemoizerTryClosed(t *testing.T) {
	t.Parallel()

	// given
	p, f := promise.New[int]()
	close(p)

	m := f.Memoize()

	ctx := context.Background()

	// when
	_, err1 := m.Try()
	_, err2 := m.Await(ctx)

	// then
	assert.ErrorIs(t, err1, promise.ErrNoResult)
	assert.ErrorIs(t, err2, promise.ErrNoResult)
}
