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
)

// A Memoizer is created with [Future.Memoize] and contains a memoized result of a future.
type Memoizer[R any] struct {
	_      noCopy
	result Result[R]
	future Future[R]
	wait   chan struct{}
}

// NewMemoizer creates a new memoizer from a future.
func NewMemoizer[R any](f <-chan Result[R]) *Memoizer[R] {
	wait := make(chan struct{}, 1)
	wait <- struct{}{}

	return &Memoizer[R]{
		wait:   wait,
		future: f,
	}
}

// Await blocks until the future is ready and returns the result.
func (m *Memoizer[R]) Await(ctx context.Context) (R, error) {
	select {
	case _, ok := <-m.wait:
		if !ok {
			return m.result.Value, m.result.Err
		}

	case <-ctx.Done():
		return *new(R), fmt.Errorf("memoizer canceled: %w", context.Cause(ctx))
	}

	var ok bool
	select {
	case m.result, ok = <-m.future:
		if !ok {
			m.result.Err = ErrNoResult
		}
		close(m.wait)

		return m.result.Value, m.result.Err

	case <-ctx.Done():
		m.wait <- struct{}{}

		return *new(R), fmt.Errorf("memoizer canceled: %w", context.Cause(ctx))
	}
}

// Try returns the result of the future if it is ready, otherwise it returns [ErrNoResult].
func (m *Memoizer[R]) Try() (R, error) {
	select {
	case _, ok := <-m.wait:
		if !ok {
			return m.result.Value, m.result.Err
		}

	default:
		return *new(R), ErrNotReady
	}

	var ok bool
	select {
	case m.result, ok = <-m.future:
		if !ok {
			m.result.Err = ErrNoResult
		}
		close(m.wait)

		return m.result.Value, m.result.Err

	default:
		m.wait <- struct{}{}

		return *new(R), ErrNotReady
	}
}
