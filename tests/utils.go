package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq" // this is actually a test package
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	rootURL        = os.Getenv("ORDERS_SERVICE_URL")
)

// Request -- как tryRequest, но падает при неуспехе.
func Request(t *testing.T, method, url string, headers http.Header, in, out interface{}) {
	t.Helper()
	err := TryRequest(t, method, url, headers, in, out)
	require.NoError(t, err)
}

func TryRequest(t *testing.T, method, url string, headers http.Header, in, out interface{}) error {
	t.Helper()
	t.Logf("%s %s with %T body", method, url, in)

	if !strings.HasPrefix(url, "http:") {
		url = rootURL + url
	}

	var body io.Reader
	if inReader, ok := in.(io.Reader); ok {
		body = inReader
	} else if in != nil {
		inJSON, err := json.Marshal(in)
		require.NoError(t, err)
		body = bytes.NewReader(inJSON)
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	for headerKey, headerVal := range headers {
		req.Header[headerKey] = headerVal
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return UnsuccessfulResponse{
			Response: resp,
			Payload:  string(outData),
		}
	}
	if out != nil {
		// Сперва зануляем out, иначе в нём может остаться что-то с предыдущего запроса.
		outv := reflect.ValueOf(out).Elem()
		outv.Set(reflect.Zero(outv.Type()))
		if err := json.Unmarshal(outData, out); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

type UnsuccessfulResponse struct {
	*http.Response
	Payload string
}

func (r UnsuccessfulResponse) Error() string {
	return fmt.Sprintf("%s: %.1000s", r.Status, r.Payload)
}
