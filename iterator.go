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
	"runtime/trace"
)

// This iterator is used to combine the results of multiple asynchronous operations waiting in parallel.
type iterator[R any] struct {
	_            noCopy
	ctx          context.Context //nolint:containedctx
	convertValue func(recv reflect.Value, ok bool) Result[R]
	cases        []reflect.SelectCase
	numFutures   int
}

func newIterator[R any, F AnyFuture](
	ctx context.Context, convertValue func(recv reflect.Value, ok bool) Result[R], l []F,
) func(yield func(int, Result[R]) bool) {
	numFutures := len(l)
	cases := make([]reflect.SelectCase, numFutures+1)
	for idx, future := range l {
		cases[idx] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: future.reflect(),
		}
	}
	cases[numFutures] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}

	i := iterator[R]{
		numFutures:   numFutures,
		cases:        cases,
		convertValue: convertValue,
		ctx:          ctx,
	}

	return i.yieldTo
}

func (i *iterator[R]) yieldTo(yield func(int, Result[R]) bool) {
	defer trace.StartRegion(i.ctx, "promiseSeq").End()
	for run := 0; run < i.numFutures; run++ {
		chosen, recv, ok := reflect.Select(i.cases)

		if chosen == i.numFutures { // context channel
			err := fmt.Errorf("list yield canceled: %w", context.Cause(i.ctx))
			i.yieldErr(yield, err)

			break
		}

		i.cases[chosen].Chan = reflect.Value{} // Disable case
		v := i.convertValue(recv, ok)
		if !yield(chosen, v) {
			break
		}
	}
}

func convertValue[R any](recv reflect.Value, ok bool) Result[R] {
	if ok {
		if r, ok2 := recv.Interface().(Result[R]); ok2 {
			return r
		}
	}

	return Result[R]{Err: ErrNoResult}
}

func convertValueAny(recv reflect.Value, ok bool) Result[any] {
	if ok {
		if a, ok2 := recv.Interface().(interface{ Any() Result[any] }); ok2 {
			return a.Any()
		}
	}

	return Result[any]{Err: ErrNoResult}
}

func (i *iterator[R]) yieldErr(yield func(int, Result[R]) bool, err error) {
	e := Result[R]{Err: err}
	for idx := 0; idx < i.numFutures; idx++ {
		if i.cases[idx].Chan.IsValid() && !yield(idx, e) {
			break
		}
	}
}
