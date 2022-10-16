package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"strings"

	"github.com/jdotw/go-utils/log"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type OPAClient interface {
	Query(ctx context.Context, query string, data interface{}, response interface{}) error
}

type opaClient struct {
	logger  log.Factory
	tracer  opentracing.Tracer
	baseURL string
}

func NewOPAClient(logger log.Factory, tracer opentracing.Tracer, baseURL string) OPAClient {
	var c OPAClient

	c = &opaClient{
		logger:  logger,
		tracer:  tracer,
		baseURL: baseURL,
	}

	return c
}

type QueryRequest struct {
	Input *interface{} `json:"input"`
}

func (c *opaClient) Query(ctx context.Context, query string, data interface{}, response interface{}) error {

	clientTrace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			c.logger.Bg().Info("connection re-use", zap.Bool("reused", info.Reused))
		},
	}
	traceCtx := httptrace.WithClientTrace(context.Background(), clientTrace)

	jsonStr, err := json.Marshal(&QueryRequest{Input: &data})
	if err != nil {
		c.logger.For(ctx).Error("Failed to marshall request body", zap.Error(err))
		return err
	}

	queryPath := strings.ReplaceAll(query, ".", "/")

	req, err := http.NewRequestWithContext(traceCtx, http.MethodPost, c.baseURL+"/v1/"+queryPath, bytes.NewBuffer(jsonStr))
	if err != nil {
		c.logger.For(ctx).Error("Failed to create request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.For(ctx).Error("Failed to perform request", zap.Error(err))
		return err
	}

	responseData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.logger.For(ctx).Error("Failed to read response body", zap.Error(err))
		return err
	}

	err = json.Unmarshal(responseData, response)
	if err != nil {
		c.logger.For(ctx).Error("Failed to unmarshal response object", zap.Error(err))
		return err
	}

	return nil
}
