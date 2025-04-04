package thirdparty

import "context"

type ThirdParty[Request any, Response any] interface {
	Handler(ctx context.Context, r *Request) (*Response, error)
}
