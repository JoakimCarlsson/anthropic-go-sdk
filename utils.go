package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// setJSONBody sets the JSON body of a request
func setJSONBody(req *http.Request, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
	req.ContentLength = int64(len(jsonBody))
	return nil
}
