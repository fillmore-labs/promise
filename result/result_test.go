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

package result_test

import (
	"errors"
	"testing"

	"fillmore-labs.com/promise/result"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("test error")

func TestV(t *testing.T) {
	t.Parallel()
	// given
	r := result.OfValue(1)
	// when
	v, err := r.V()
	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 1, v)
	}
}

func TestVErr(t *testing.T) {
	t.Parallel()
	// given
	r := result.OfError[struct{}](errTest)
	// when
	_, err := r.V()
	// then
	assert.ErrorIs(t, err, errTest)
}

func TestOf(t *testing.T) {
	t.Parallel()
	// given
	r := result.Of(1, nil)
	// when
	v := r.Value()
	err := r.Err()
	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 1, v)
	}
}

func TestOfErr(t *testing.T) {
	t.Parallel()
	// given
	r := result.Of(1, errTest)
	// when
	_ = r.Value() // doesn't panic
	err := r.Err()
	// then
	assert.ErrorIs(t, err, errTest)
}

func TestAny(t *testing.T) {
	t.Parallel()
	// given
	r := result.OfValue(1)
	// when
	r2 := r.Any()
	// then
	if assert.NoError(t, r2.Err()) {
		assert.Equal(t, 1, r2.Value())
	}
}

func TestAnyErr(t *testing.T) {
	t.Parallel()
	// given
	r := result.OfError[int](errTest)
	// when
	r2 := r.Any()
	// then
	assert.ErrorIs(t, r2.Err(), errTest)
	_ = r2.Value()
}
