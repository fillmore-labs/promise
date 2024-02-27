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

//go:build goexperiment.rangefunc

package promise_test

import (
	"context"
	"testing"

	"fillmore-labs.com/promise"
	"fillmore-labs.com/promise/result"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	values := []struct {
		value int
		err   error
	}{
		{1, nil},
		{0, errTest},
		{2, nil},
	}

	for i, v := range values {
		value, err := v.value, v.err
		go promises[i].Do(func() (int, error) { return value, err })
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	var results [3]result.Result[int]
	for i, r := range futures.All(ctx) { //nolint:typecheck
		results[i] = r
	}

	// then
	if assert.NoError(t, results[0].Err()) {
		assert.Equal(t, 1, results[0].Value())
	}
	if assert.ErrorIs(t, results[1].Err(), errTest) {
		_ = results[1].Value() // Should not panic
	}
	if assert.NoError(t, results[2].Err()) {
		assert.Equal(t, 2, results[2].Value())
	}
}

func TestAllEmpty(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var futures promise.List[int]

	// when
	allFutures := futures.All(ctx)

	// then
	for _, v := range allFutures { //nolint:typecheck
		t.Errorf("Invalid value %v", v)
	}
}
