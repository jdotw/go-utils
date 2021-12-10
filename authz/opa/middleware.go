package opa

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/12kmps/baas/log"
	"github.com/12kmps/baas/opa"
	"github.com/12kmps/baas/tracing"
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

			results, err := query.Eval(ctx, rego.EvalInput(request))
			if err != nil {
				// handle error
				return nil, err
			}

			if !results.Allowed() {
				a.logger.For(ctx).Info("Denied by policy", zap.String("query", queryString))
				return nil, errors.New("Denied by policy")
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
	c := opa.NewOPAClient(a.logger, a.tracer, "http://localhost:8181")
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, span := tracing.NewChildSpanAndContext(ctx, a.tracer, "AuthZPolicyExternal")

			b, err := json.Marshal(request)
			if err != nil {
				// handle error
				return nil, err
			}

			var resp AuthorizationResponse
			err = c.Query(ctx, queryString, b, &resp)
			if err != nil {
				// handle error
				return nil, err
			}

			if !resp.Result {
				a.logger.For(ctx).Info("Denied by policy", zap.String("query", queryString))
				return nil, errors.New("Denied by policy")
			}

			ctx = context.WithValue(ctx, AuthorizationResultsContextKey, resp)
			span.Finish()

			return next(ctx, request)
		}
	}

}
