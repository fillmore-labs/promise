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

package result

// Result defines the interface for returning results from asynchronous operations.
// It encapsulates the final value or error from the operation.
type Result[R any] interface {
	AnyResult
	V() (R, error) // The V method returns the final value or an error.
	Value() R      // The Value method returns the final value.
	Err() error    // The Err method returns the error.
}

// AnyResult can be used with any [Result].
type AnyResult interface {
	Any() Result[any] // The Any method returns a Result[any] that can be used with any type.
}

// Of creates a new [Result] from a pair of values.
func Of[R any](value R, err error) Result[R] {
	if err != nil {
		return errorResult[R]{err: err}
	}

	return valueResult[R]{value: value}
}

// OfValue creates a new [Result] from a value.
func OfValue[R any](value R) Result[R] {
	return valueResult[R]{value: value}
}

// OfError creates a new [Result] from an error.
func OfError[R any](err error) Result[R] {
	return errorResult[R]{err: err}
}

// valueResult is an implementation of [Result] that simply holds a value.
type valueResult[R any] struct {
	value R
}

// V returns the stored value.
func (v valueResult[R]) V() (R, error) {
	return v.value, nil
}

// Value returns the stored value.
func (v valueResult[R]) Value() R {
	return v.value
}

// The Err method returns nil.
func (v valueResult[_]) Err() error {
	return nil
}

// Any returns the valueResult as a Result[any].
func (v valueResult[_]) Any() Result[any] {
	return valueResult[any]{value: v.value}
}

// errorResult handles errors from failed operations.
type errorResult[_ any] struct {
	err error
}

// V returns the stored error.
func (e errorResult[R]) V() (R, error) {
	return *new(R), e.err
}

// Value returns the null value.
func (e errorResult[R]) Value() R {
	return *new(R)
}

// Err returns the stored error.
func (e errorResult[_]) Err() error {
	return e.err
}

// Any returns the errorResult as a Result[any].
func (e errorResult[_]) Any() Result[any] {
	return errorResult[any](e)
}
