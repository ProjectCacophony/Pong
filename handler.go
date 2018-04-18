package main

import (
	"context"

	"github.com/json-iterator/go"

	"gitlab.com/project-d-collab/dhelpers"
)

// lambdaHandler is the generic function type
type lambdaHandler func(context.Context, []byte) ([]byte, error)

func (handler lambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	_, err := handler(ctx, payload)
	return nil, err
}

func newHandler(_ interface{}) lambdaHandler {
	return func(ctx context.Context, payload []byte) ([]byte, error) {
		var container dhelpers.EventContainer
		err := jsoniter.Unmarshal(payload, &container)
		if err != nil {
			return nil, err
		}

		err = Handler(container)
		return nil, err
	}
}
