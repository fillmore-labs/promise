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

package group_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"fillmore-labs.com/promise/group"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

var errTest = errors.New("test error")

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDoAsync(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancelCause(context.Background())
	g := group.New(group.WithCancel(cancel), group.WithLimit(1))

	// when
	f := group.DoAsync(ctx, g, func() (int, error) { return 0, errTest })
	err := g.Wait()
	_, errf := f.Try()
	cause := context.Cause(ctx)

	// then
	assert.ErrorIs(t, err, errTest)
	assert.ErrorIs(t, errf, errTest)
	assert.ErrorIs(t, cause, errTest)

	select {
	case <-ctx.Done():
	default:
		assert.Fail(t, "context should be canceled")
	}
}

func TestGroupReject(t *testing.T) {
	t.Parallel()

	// given
	g := group.New(group.WithLimit(1))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	ch := make(chan int)
	f1 := group.DoAsync(ctx, g, func() (int, error) {
		return <-ch, nil
	})

	ctx2, cancel2 := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel2()

	f2 := group.DoAsync(ctx2, g, func() (int, error) { return 1, nil })
	ch <- 1

	_, err2 := f2.Try()
	err := g.Wait()
	v1, err1 := f1.Try()

	// then
	assert.NoError(t, err)
	if assert.NoError(t, err1) {
		assert.Equal(t, 1, v1)
	}
	assert.ErrorIs(t, err2, context.DeadlineExceeded)
}

func TestGo(t *testing.T) {
	t.Parallel()

	// given
	var g group.Group

	// when
	g.Go(func() error { return errTest })

	err := g.Wait()

	// then
	assert.ErrorIs(t, err, errTest)
}

func TestGoReject(t *testing.T) {
	t.Parallel()

	// given
	g := group.New(group.WithLimit(1))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	ch := make(chan error)
	g.GoCtx(ctx, func() error {
		return <-ch
	})

	ctx2, cancel2 := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel2()

	g.GoCtx(ctx2, func() error { return nil })
	ch <- errTest

	err := g.Wait()

	// then
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestGroupLimit(t *testing.T) {
	t.Parallel()

	// given
	defer func() { _ = recover() }()

	// when
	_ = group.New(group.WithLimit(0))

	// then
	assert.Fail(t, "limit 0 should panic")
}

func TestGroupPanic(t *testing.T) {
	t.Parallel()

	// given
	const msg = "test"
	var g group.Group

	// when
	f := group.DoAsync(context.Background(), &g, func() (int, error) { panic(msg) })

	var p any
	func() {
		defer func() { p = recover() }()
		_ = g.Wait()
	}()

	_, errf := f.Try()

	// then

	assert.ErrorContains(t, errf, msg)
	var err group.ExecutionError
	if assert.ErrorAs(t, errf, &err) {
		assert.Equal(t, msg, err.Value)
	}
	assert.Equal(t, msg, p)
}
