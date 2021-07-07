/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package request

import (
	"math"
	"net/http"

	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

func newMutatingWidthEstimator(countFn watchCountGetterFunc) WidthEstimatorFunc {
	estimator := &mutatingWidthEstimator{
		countFn: countFn,
	}
	return estimator.estimate
}

type mutatingWidthEstimator struct {
	countFn watchCountGetterFunc
}

func (e *mutatingWidthEstimator) estimate(r *http.Request) Width {
	requestInfo, ok := apirequest.RequestInfoFrom(r.Context())
	if !ok {
		return Width{Seats: maximumSeats}
	}

	watchCount := e.countFn.Get(requestInfo)

	// for now, our rough estimate is to allocate one seat for each each 100 watchers
	// potentially interested in a given object.
	//
	// TODO: As described in the KEP it should be much more sophisticated, including:
	// - taking advantage of `additional latency` concept once this is implemented
	// - taking into account cost of a single event (different events may have
	//   different size).
	// However, we start simple first to get some operational experience from it.
	seats := uint(math.Ceil(float64(watchCount) / float64(100)))
	if seats < minimumSeats {
		seats = minimumSeats
	}
	if seats > maximumSeats {
		seats = maximumSeats
	}
	return Width{Seats: seats}
}
