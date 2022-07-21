package tests

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq" // this is actually a test package
	"github.com/stretchr/testify/assert"
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

// TryExportRequest делает HTTP-запрос method на url, передавая in в качестве тела (если не nil)
// и возвращая ответ в out (если не nil).
func TryExportRequest(t *testing.T, url string, headers http.Header, out *[][]string) error {
	t.Helper()
	t.Logf("%s %s", "GET", url)

	if !strings.HasPrefix(url, "http:") {
		url = rootURL + url
	}

	var body io.Reader
	req, err := http.NewRequest("GET", url, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/csv")

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
		r := csv.NewReader(bytes.NewReader(outData))
		var records [][]string
		records, err = r.ReadAll()
		if err != nil {
			return err
		}
		*out = records
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


// AssertJSON проверяет, что data содержит в себе pathvals -- пары path и value,
// где path -- путь внутри JSON-а (как принимает github.com/Jeffail/gabs/v2.(*Container).Path),
// а value -- ожидаемый JSON по этому пути (именно JSON: строки должны быть в кавычках и т. д.).
func AssertJSON(t *testing.T, data interface{}, pathvals ...string) bool { //nolint:unparam
	t.Helper()
	data = convertJSON(data)
	g := gabs.Wrap(data)
	for i := 0; i < len(pathvals); i += 2 {
		path := pathvals[i]
		expected := pathvals[i+1]
		actual := g.Path(path).String()
		if !assert.JSONEq(t, expected, actual, "at %q", path) {
			return false
		}
	}
	return true
}
