package transport

import (
	"errors"
	"net/http"
)

// RespondError sends a simple error response
func (t *Default) RespondError(w http.ResponseWriter, data error) {
	var apiErr *APIError
	if errors.As(data, &apiErr) {
		t.RespondJSON(w, apiErr.Code, apiErr)
		return
	}

	t.RespondJSON(w, http.StatusInternalServerError, "Internal Server Error")
}

