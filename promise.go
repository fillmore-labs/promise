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

import "fillmore-labs.com/promise/result"

// Promise is used to send the result of an asynchronous operation.
//
// It is a write-only promise.
// Either [Promise.Resolve] or [Promise.Reject] should be called exactly once.
type Promise[R any] chan<- result.Result[R]

// New provides a simple way to create a [Promise] for asynchronous operations.
// This allows synchronous and asynchronous code to be composed seamlessly and separating initiation from running.
//
// The returned [Future] can be used to retrieve the eventual result of the [Promise].
func New[R any]() (Promise[R], Future[R]) {
	ch := make(chan result.Result[R], 1)

	return ch, ch
}

// Resolve fulfills the promise with a value.
func (p Promise[R]) Resolve(value R) {
	p.complete(result.OfValue(value))
}

// Reject breaks the promise with an error.
func (p Promise[R]) Reject(err error) {
	p.complete(result.OfError[R](err))
}

// Do runs f synchronously, resolving the promise with the return value.
func (p Promise[R]) Do(f func() (R, error)) {
	p.complete(result.Of(f()))
}

func (p Promise[R]) complete(r result.Result[R]) {
	p <- r
	close(p)
}
