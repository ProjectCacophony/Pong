package main

import (
	"context"
	"reflect"

	"encoding/json"

	"github.com/json-iterator/go"
)

// lambdaHandler is the generic function type
type lambdaHandler func(context.Context, []byte) (interface{}, error)

// Invoke calls the handler, and serializes the response.
// If the underlying handler returned an error, or an error occurs during serialization, error is returned.
func (handler lambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	response, err := handler(ctx, payload)
	if err != nil {
		return nil, err
	}

	responseBytes, err := jsoniter.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

// newHandler Creates the base lambda handler, which will do basic payload unmarshaling before defering to handlerSymbol.
// If handlerSymbol is not a valid handler, the returned function will be a handler that just reports the validation error.
func newHandler(handlerSymbol interface{}) lambdaHandler {
	handler := reflect.ValueOf(handlerSymbol)
	handlerType := reflect.TypeOf(handlerSymbol)

	return func(ctx context.Context, payload []byte) (interface{}, error) {
		eventType := handlerType.In(handlerType.NumIn() - 1)
		event := reflect.New(eventType)
		// construct arguments
		var args []reflect.Value
		if err := json.Unmarshal(payload, event.Interface()); err != nil {
			return nil, err
		}

		args = append(args, event.Elem())

		response := handler.Call(args)

		// convert return values into (interface{}, error)
		var err error
		if len(response) > 0 {
			if errVal, ok := response[len(response)-1].Interface().(error); ok {
				err = errVal
			}
		}
		var val interface{}
		if len(response) > 1 {
			val = response[0].Interface()
		}

		return val, err
	}
}
