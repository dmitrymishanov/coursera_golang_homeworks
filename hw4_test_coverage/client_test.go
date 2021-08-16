package hw4_test_coverage

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const AccessToken = "access token"

type Dataset struct {
	Users []UserWithTags `xml:"row"`
}

type UserWithTags struct {
	Id     int    `xml:"id" json:"id"`
	Name   string `xml:"first_name" json:"name"`
	Age    int    `xml:"age" json:"age"`
	About  string `xml:"about" json:"about"`
	Gender string `xml:"gender" json:"gender"`
}

var dataset Dataset

func init() {
	xmlText, _ := ioutil.ReadFile("dataset.xml")
	xml.Unmarshal(xmlText, &dataset)
}

type ErrorCase struct {
	request       SearchRequest
	expectedError string
}

func SearchServerUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != AccessToken {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func TestFindUsersUnauthorized(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerUnauthorizedHandler))
	defer testServer.Close()
	client := SearchClient{"foo bar", testServer.URL}

	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "Bad AccessToken" {
		t.Errorf("Expected unauthorized error")
	}
}

func SearchServerInternalErrorHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestFindUsersInternalServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerInternalErrorHandler))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}

	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "SearchServer fatal error" {
		t.Errorf("Expected 500")
	}
}

func SearchServerBadOrderFieldHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	allowedOrderFields := map[string]bool{"Id": true, "Age": true, "Name": true}
	q := r.URL.Query()
	orderField := q["order_field"]
	if len(orderField) > 0 {
		if allowed := allowedOrderFields[orderField[0]]; !allowed {
			_, err := io.WriteString(w, `{"Error": "ErrorBadOrderField"}`)
			if err != nil {
				panic("UnexpectedError")
			}
		}
	}
}

func TestFindUsersBadOrderField(t *testing.T) {
	ErrorExpected := "OrderField foo bar baz invalid"
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerBadOrderFieldHandler))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}

	_, err := client.FindUsers(SearchRequest{OrderField: "foo bar baz"})
	if err == nil || err.Error() != ErrorExpected {
		t.Errorf("Expected BadRequest because of %s", ErrorExpected)
	}
}

func SearchServerBadRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func TestFindUsersBadRequest(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerBadRequest))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}

	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "cant unpack error json: unexpected end of JSON input" {
		t.Errorf("Expected problems with unpacking json")
	}
}

func SearchServerBadRequestUnknown(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := io.WriteString(w, `{"Error": "SomeError"}`)
	if err != nil {
		panic("Unexpected error")
	}
}

func TestFindUsersBadBodyReturned(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerBadRequestUnknown))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}

	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "unknown bad request error: SomeError" {
		t.Errorf("Expected to catch unknown error")
	}
}

func TestFindUsersNegativeValues(t *testing.T) {
	cases := []ErrorCase{
		{SearchRequest{Limit: -1}, "limit must be > 0"},
		{SearchRequest{Offset: -1}, "offset must be > 0"},
	}

	client := SearchClient{"", ""}

	for _, c := range cases {
		_, err := client.FindUsers(c.request)
		if err == nil || err.Error() != c.expectedError {
			t.Errorf("Error expected")
		}
	}
}

func SearchServerWithTimeout(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(time.Second * 1)
	w.WriteHeader(http.StatusOK)
}

func TestFindUsers(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerWithTimeout))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}
	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "timeout for limit=1&offset=0&order_by=0&order_field=&query=" {
		t.Errorf("Timeout expected")
	}
}

func TestFindUsersBadError(t *testing.T) {
	client := SearchClient{AccessToken, "bad_url"}
	_, err := client.FindUsers(SearchRequest{})

	if err == nil || !strings.HasPrefix(err.Error(), "unknown error") {
		t.Errorf(err.Error())
	}
}

func SearchServerInvalidJsonResp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `Invalid json string`)
}

func TestFindUsersInvalidResult(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerInvalidJsonResp))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}
	_, err := client.FindUsers(SearchRequest{})

	if !strings.HasPrefix(err.Error(), "cant unpack result json") {
		t.Errorf("Did't get expected error for invalid json.")
	}
}

func SearchServerOk(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))

	if limit > 25 {
		limit = 25
	}
	start := offset
	if start > len(dataset.Users) {
		start = len(dataset.Users)
	}
	end := offset + limit
	if end > len(dataset.Users) {
		end = len(dataset.Users)
	}
	body, _ := json.Marshal(dataset.Users[start:end])
	w.Write(body)
}

type TestCase struct {
	limit             int
	expectedResultLen int
	nextPage          bool
}

func TestFindUsersOK(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServerOk))
	defer testServer.Close()
	client := SearchClient{AccessToken, testServer.URL}

	cases := []TestCase{
		{5, 5, true},
		{50, 25, false},
	}

	for _, tc := range cases {
		result, err := client.FindUsers(SearchRequest{Limit: tc.limit})
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if len(result.Users) != tc.expectedResultLen {
			t.Errorf("Expected len: %d, real len: %d", tc.expectedResultLen, len(result.Users))
		}
		if result.NextPage != tc.nextPage {
			t.Errorf("Wrong nextPage value: %t", result.NextPage)
		}
	}
}
