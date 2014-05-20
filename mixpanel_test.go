package mixpanel

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

var (
	ts          *httptest.Server
	client      *Mixpanel
	LastRequest *http.Request
)

func setup() {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		LastRequest = r
	}))

	client = NewMixpanel("e3bc4100330c35722740fb8c6f5abddc")
	client.ApiURL = ts.URL
}

func teardown() {
	ts.Close()
}

func decodeURL(url string) string {
	data := strings.Split(url, "data=")[1]
	decoded, _ := base64.StdEncoding.DecodeString(data)
	return string(decoded[:])
}

// examples from https://mixpanel.com/help/reference/http

func TestTrack(t *testing.T) {
	setup()
	defer teardown()

	client.Track("13793", "Signed Up", map[string]interface{}{
		"Referred By": "Friend",
	})

	want := "{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/track"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestIdentify(t *testing.T) {
	setup()
	defer teardown()

	client.Identify("13793")

	want := "{\"$distinct_id\":\"13793\",\"$token\":\"e3bc4100330c35722740fb8c6f5abddc\"}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/engage"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestPeopleOperations(t *testing.T) {
	setup()
	defer teardown()

	people := client.Identify("13793")
	people.Update("$set", map[string]interface{}{
		"Address":  "1313 Mockingbird Lane",
		"Birthday": "1948-01-01",
	})

	want := "{\"$distinct_id\":\"13793\",\"$set\":{\"Address\":\"1313 Mockingbird Lane\",\"Birthday\":\"1948-01-01\"},\"$token\":\"e3bc4100330c35722740fb8c6f5abddc\"}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/engage"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestPeopleTrack(t *testing.T) {
	setup()
	defer teardown()

	people := client.Identify("13793")
	people.Track("Signed Up", map[string]interface{}{
		"Referred By": "Friend",
	})

	want := "{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/track"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}
