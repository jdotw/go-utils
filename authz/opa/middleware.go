package opa

import (
	"context"
	"os"

	"github.com/jdotw/go-utils/authn/jwt"
	"github.com/jdotw/go-utils/authzerrors"
	"github.com/jdotw/go-utils/log"
	"github.com/jdotw/go-utils/opa"
	"github.com/jdotw/go-utils/tracing"
	"github.com/go-kit/kit/endpoint"
	"github.com/open-policy-agent/opa/rego"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type contextKey string

const (
	// JWTContextKey holds the key used to store a JWT in the context.
	AuthorizationResultsContextKey contextKey = "AuthZResults"
)

type Authorizor struct {
	logger log.Factory
	tracer opentracing.Tracer
	query  rego.PreparedEvalQuery
}

func NewAuthorizor(logger log.Factory, tracer opentracing.Tracer) Authorizor {
	return Authorizor{
		logger: logger,
		tracer: tracer,
	}
}

type queryInput struct {
	Request interface{} `json:"request,omitempty"`
	Claims  interface{} `json:"claims,omitempty"`
}

func inputForRequest(ctx context.Context, request interface{}) queryInput {
	return queryInput{
		Request: request,
		Claims:  ctx.Value(jwt.JWTClaimsContextKey),
	}
}

func (a *Authorizor) NewInProcessMiddleware(policy string, queryString string) endpoint.Middleware {
	query, err := rego.New(
		rego.Query(queryString),
		rego.Module("policy.rego", policy),
	).PrepareForEval(context.Background())
	if err != nil {
		a.logger.Bg().Fatal("Failed to prepare endpoint authorization policy", zap.Error(err))
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, span := tracing.NewChildSpanAndContext(ctx, a.tracer, "AuthZPolicyInternal")

			results, err := query.Eval(ctx, rego.EvalInput(inputForRequest(ctx, request)))
			if err != nil {
				// handle error
				return nil, err
			}

			if !results.Allowed() {
				a.logger.For(ctx).Info("Denied by policy", zap.String("query", queryString))
				return nil, authzerrors.ErrDeniedByPolicy
			}

			ctx = context.WithValue(ctx, AuthorizationResultsContextKey, results)
			span.Finish()

			return next(ctx, request)
		}
	}
}

type AuthorizationResponse struct {
	Result bool `json:"result,omitempty"`
}

func (a *Authorizor) NewSidecarMiddleware(queryString string) endpoint.Middleware {
	h := os.Getenv("OPA_HOST")
	if len(h) == 0 {
		h = "localhost"
	}
	p := os.Getenv("OPA_PORT")
	if len(p) == 0 {
		p = "8181"
	}
	c := opa.NewOPAClient(a.logger, a.tracer, "http://"+h+":"+p)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, span := tracing.NewChildSpanAndContext(ctx, a.tracer, "AuthZPolicyExternal")

			var resp AuthorizationResponse
			err = c.Query(ctx, queryString, inputForRequest(ctx, request), &resp)
			if err != nil {
				// handle error
				return nil, err
			}

			if !resp.Result {
				a.logger.For(ctx).Info("Denied by policy", zap.String("query", queryString))
				return nil, authzerrors.ErrDeniedByPolicy
			}

			ctx = context.WithValue(ctx, AuthorizationResultsContextKey, resp)
			span.Finish()

			return next(ctx, request)
		}
	}

}
