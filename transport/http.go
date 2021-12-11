package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/12kmps/baas/authzerrors"
	"github.com/12kmps/baas/recorderrors"
)

// Response Encoder (Generic)

func HTTPEncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// Error Encoder

type HTTPErrorResponse struct {
	Error string `json:"error,omitempty"`
}

func HTTPErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	if err == recorderrors.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
	} else if err == authzerrors.ErrDeniedByPolicy {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HTTPErrorResponse{Error: err.Error()})
	}
}
