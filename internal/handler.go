package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"io"
	"net/http"
)

func ParseAndValidateRequestBody(ctx context.Context, v *validator.Validate, r *http.Request, s interface{}) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			fmt.Println("Error closing request body:", err)
		}
	}()

	err = json.Unmarshal(bodyBytes, s)
	if err != nil {
		return err
	}

	err = v.Struct(s)
	if err != nil {
		return err
	}

	return nil
}

func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func ParseUUID(value string) (uuid.UUID, error) {
	parsedUUID, err := uuid.Parse(value)
	if err != nil {
		return parsedUUID, fmt.Errorf("failed to parse UUID: %w", err)
	}
	return parsedUUID, nil
}
