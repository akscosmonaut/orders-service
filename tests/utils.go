package tests

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // this is actually a test package
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

)

var (
	rootURL        = os.Getenv("ORDERS_SERVICE_URL")
)

const (
	PickerSubject  = "picker"
	AdminSubject   = "admin"
	AdminPrefix    = AdminSubject + "_"
	ZonelessFormat = "2006-01-02T15:04:05"
	DateFormat     = "2006-01-02"
)

type MockInterface interface {
	On(methodName string, arguments ...interface{}) *mock.Call
}

// Request -- как tryRequest, но падает при неуспехе.
func Request(t *testing.T, method, url string, headers http.Header, in, out interface{}) {
	t.Helper()
	err := TryRequest(t, method, url, headers, in, out)
	require.NoError(t, err)
}

// TryRequest делает HTTP-запрос method на url, передавая in в качестве тела (если не nil)
// и десериализуя тело ответа в out (если не nil).
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

// ExportRequest -- как tryExportRequest, но падает при неуспехе.
func ExportRequest(t *testing.T, url string, headers http.Header, out *[][]string) {
	t.Helper()
	err := TryExportRequest(t, url, headers, out)
	require.NoError(t, err)
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

func PickerAuth(t *testing.T, phone string) http.Header {
	var resp map[string]interface{}
	Request(t, "POST", "app/v2/auth/sms/verify", nil, json.RawMessage(`{
			"phone": "`+phone+`"
		}`), &resp)
	token := generateAuthToken(PickerSubject, phone, resp["fingerprint"].(string))
	return http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
}

func Unauthorized() http.Header {
	return http.Header{
		"Authorization": {"Bearer wrong"},
	}
}

func getAuthToken(phone, fingerprint string) http.Header {
	headers := http.Header{}
	var signedToken string

	if phone == "" || fingerprint == "" {
		return headers
	}

	if _, alreadyHaveToken := userTokenMap[phone]; alreadyHaveToken {
		signedToken = userTokenMap[phone]
	} else {
		signedToken = generateAuthToken(AdminSubject, phone, fingerprint)
		userTokenMap[phone] = signedToken
	}

	headers = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", signedToken)},
	}

	return headers
}
func generateAuthToken(subject, phone, fingerprint string) string {
	var err error

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        uuid.New().String(),
		Audience:  fingerprint,
		Issuer:    phone,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		Subject:   subject,
	},
	)

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		panic(err)
	}

	if subject == AdminSubject {
		signedToken = AdminPrefix + signedToken
	}

	return signedToken
}

func AdminAuth() http.Header {
	return getAuthToken("+75554440001", "b1e4d643-fe26-48a5-bab6-0298e85418e0")
}

func ManagerAuth() http.Header {
	return getAuthToken("+75554440003", "d2f27d71-73be-41cc-8a1b-85f48ffc7d5b")
}

func OperatorAuth() http.Header {
	return getAuthToken("+75554440004", "92953d74-cf40-4346-a7fe-5dfb2eea878c")
}

func ManagerMultipleNetworksLogServicesAuth() http.Header {
	return getAuthToken("+75554479805", "b67c718f-fa89-4824-8028-dbd7aadf429f")
}

func OperatorMultipleNetworksLogServicesAuth() http.Header {
	return getAuthToken("+75554479806", "82822937-45b0-4bf8-95e3-755adfd1673b")
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

func ExtractJSON(data interface{}, path string) interface{} {
	data = convertJSON(data)
	return gabs.Wrap(data).Path(path).Data()
}

// convertJSON преобразует json.RawMessage в вид, пригодный для gabs.Wrap.
func convertJSON(data interface{}) interface{} {
	if raw, ok := data.(json.RawMessage); ok {
		var parsed interface{}
		err := json.Unmarshal([]byte(raw), &parsed)
		if err != nil {
			panic(err)
		}
		data = parsed
	}
	return data
}

func AsJSON(v interface{}) string {
	j, err := json.Marshal(v)
	if err != nil {
		panic("cannot marshal to JSON: " + err.Error())
	}
	return string(j)
}

// GetMockJournal возвращает журнал запросов к Wiremock-у, опционально отфильтрованный по criteria.
func GetMockJournal(t *testing.T, criteria interface{}) []interface{} {
	if criteria == nil {
		criteria = make(map[string]interface{})
	}
	var r struct {
		Requests []interface{}
	}
	Request(t, "POST", mockURL+"__admin/requests/find", nil, criteria, &r)
	return r.Requests
}

// ResetMockJournal очищает журнал запросов к Wiremock-у.
func ResetMockJournal(t *testing.T) {
	Request(t, "DELETE", mockURL+"__admin/requests", nil, nil, nil)
}

func ParseTime(t *testing.T, v interface{}) time.Time {
	vt, err := time.Parse(time.RFC3339, v.(string))
	require.NoErrorf(t, err, "bad time %q", v)
	return vt
}

func FileForm(header http.Header, fileContents string) (http.Header, io.Reader) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("file", "file.txt")
	if err != nil {
		panic(err)
	}
	if _, err := fw.Write([]byte(fileContents)); err != nil {
		panic(err)
	}
	w.Close()
	header = header.Clone()
	header.Set("Content-Type", w.FormDataContentType())
	return header, buf
}

// metrics implements MetricRegister
type metrics bool

func (metrics) MustRegister(...prometheus.Collector) {
}

func AssertExportedColumn(t *testing.T, expected []string, actual [][]string, rowIndex int) {
	for i, item := range expected {
		assert.Equal(t, item, actual[i+1][rowIndex])
	}
}

func getLKAAuthToken(networkID int64, rights []string) http.Header {
	var signedToken string

	if _, alreadyHaveToken := lkaUserTokenMap[networkID]; alreadyHaveToken {
		signedToken = lkaUserTokenMap[networkID]
	} else {
		signedToken = generateLKAAuthToken(networkID, rights)
		lkaUserTokenMap[networkID] = signedToken
	}

	headers := http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", signedToken)},
	}

	return headers
}
func generateLKAAuthToken(networkID int64, rights []string) string {
	var err error

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        uuid.New().String(),
		Issuer:    strconv.FormatInt(networkID, 10),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		Subject:   strings.Join(rights, ","),
	},
	)

	signedToken, err := token.SignedString([]byte(lkaTokenSecret))
	if err != nil {
		panic(err)
	}

	return signedToken
}

func LkaAuth(networkID int64) http.Header {
	return getLKAAuthToken(networkID, []string{})
}

func LkaAuthWithRights(networkID int64, rights []string) http.Header {
	return getLKAAuthToken(networkID, rights)
}

func ParseFlatSchedules(data []interface{}) (entities.FlatSchedules, error) {
	var flat entities.FlatSchedules
	if len(data) != 1 {
		return flat, errors.New("len(data) != 1")
	}

	b, ok := data[0].([]byte)
	if !ok {
		return flat, errors.New("cannot assert data to bytes")
	}

	err := json.Unmarshal(b, &flat)
	if err != nil {
		return flat, err
	}
	return flat, nil
}

func CheckExistsInUniqueSliceOfStrings(s []interface{}, vals ...string) bool {
	r := make(map[string]struct{})
	for i := range s {
		r[s[i].(string)] = struct{}{}
	}
	if len(s) != len(r) {
		return false
	}

	for i := range vals {
		if _, ok := r[vals[i]]; !ok {
			return false
		}
	}
	return true

}
