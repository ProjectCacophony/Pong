package main

import (
	"context"

	"github.com/json-iterator/go"

	"fmt"
	"time"

	"gitlab.com/project-d-collab/dhelpers"
)

// lambdaHandler is the generic function type
type lambdaHandler func(context.Context, []byte) ([]byte, error)

func (handler lambdaHandler) Invoke(_ context.Context, payload []byte) ([]byte, error) {
	_, err := handler(nil, payload)
	return nil, err
}

func newHandler(_ interface{}) lambdaHandler {
	return func(_ context.Context, payload []byte) ([]byte, error) {
		unmarshalStart := time.Now()
		var container dhelpers.EventContainer
		err := jsoniter.Unmarshal(payload, &container)
		if err != nil {
			return nil, err
		}
		// benchmark
		fmt.Println("unmarshal took", time.Now().Sub(unmarshalStart).String())

		err = Handler(container)
		return nil, err
	}
}
