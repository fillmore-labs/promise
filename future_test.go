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
	"errors"
	"testing"

	"fillmore-labs.com/promise"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FutureTestSuite struct {
	suite.Suite
	promise promise.Promise[int]
	future  promise.Future[int]
}

func TestAsyncTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(FutureTestSuite))
}

func (s *FutureTestSuite) SetupTest() {
	s.promise, s.future = promise.New[int]()
}

func (s *FutureTestSuite) TestValue() {
	// given
	s.promise.Do(func() (int, error) { return 1, nil })

	// when
	value, err := s.future.Try()

	// then
	if s.NoError(err) {
		s.Equal(1, value)
	}
}

var errTest = errors.New("test error")

func (s *FutureTestSuite) TestError() {
	// given
	s.promise.Do(func() (int, error) { return 0, errTest })

	// when
	_, err := s.future.Try()

	// then
	s.ErrorIs(err, errTest)
}

func (s *FutureTestSuite) TestCancellation() {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when
	_, err := s.future.Await(ctx)

	// then
	s.ErrorIs(err, context.Canceled)
}

func (s *FutureTestSuite) TestNoResult() {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	close(s.promise)

	// when
	_, err := s.future.Await(ctx)

	// then
	s.ErrorIs(err, promise.ErrNoResult)
}

func (s *FutureTestSuite) TestTry() {
	// given

	// when
	_, err1 := s.future.Try()

	s.promise.Resolve(1)

	v2, err2 := s.future.Try()
	_, err3 := s.future.Try()

	// then
	s.ErrorIs(err1, promise.ErrNotReady)
	if s.NoError(err2) {
		s.Equal(1, v2)
	}
	s.ErrorIs(err3, promise.ErrNoResult)
}

func TestAsync(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	f := promise.NewAsync(func() (int, error) { return 1, nil })
	value, err := f.Await(ctx)

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 1, value)
	}
}

func TestAsyncCancel(t *testing.T) { //nolint:paralleltest
	// given
	ctx := context.Background()
	ctx1, cancel1 := context.WithCancel(ctx)

	// when
	_, f := promise.New[int]()
	ff := promise.NewAsync(func() (int, error) { return f.Await(ctx1) })
	_, err1 := ff.Try()

	cancel1()

	_, err2 := ff.Await(ctx)

	// then
	assert.ErrorIs(t, err1, promise.ErrNotReady)
	assert.ErrorIs(t, err2, context.Canceled)
}

func TestNil(t *testing.T) {
	t.Parallel()

	// given
	p := promise.Promise[int](nil)

	// when
	p.Resolve(1)

	// then
}
