package mixpanel

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	ts          *httptest.Server
	client      Mixpanel
	LastRequest *http.Request
)

func setup() {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("1\n"))
		LastRequest = r
	}))

	client = NewWithSecret("e3bc4100330c35722740fb8c6f5abddc", "mysecret", ts.URL)
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

	client.Track("13793", "Signed Up", &Event{
		Properties: map[string]interface{}{
			"Referred By": "Friend",
		},
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

func TestImport(t *testing.T) {
	setup()
	defer teardown()

	importTime := time.Now().Add(-5 * 24 * time.Hour)

	client.Import("13793", "Signed Up", &Event{
		Properties: map[string]interface{}{
			"Referred By": "Friend",
		},
		Timestamp: &importTime,
	})

	want := fmt.Sprintf("{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"time\":%d,\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}", importTime.Unix())

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/import"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestGroupOperations(t *testing.T) {
	setup()
	defer teardown()

	client.UpdateGroup("company_id", "11", &Update{
		Operation: "$set",
		Properties: map[string]interface{}{
			"Address":  "1313 Mockingbird Lane",
			"Birthday": "1948-01-01",
		},
	})

	want := "{\"$group_id\":\"11\",\"$group_key\":\"company_id\",\"$set\":{\"Address\":\"1313 Mockingbird Lane\",\"Birthday\":\"1948-01-01\"},\"$token\":\"e3bc4100330c35722740fb8c6f5abddc\"}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/groups"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestUpdate(t *testing.T) {
	setup()
	defer teardown()

	client.Update("13793", &Update{
		Operation: "$set",
		Properties: map[string]interface{}{
			"Address":  "1313 Mockingbird Lane",
			"Birthday": "1948-01-01",
		},
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

func TestError(t *testing.T) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"error": "some error", "status": 0}`))
		LastRequest = r
	}))

	assertErrTrackFailed := func(err error) {
		merr, ok := err.(*MixpanelError)

		if !ok {
			t.Errorf("Error should be wrapped in a MixpanelError: %v", err)
			return
		}

		terr, ok := merr.Err.(*ErrTrackFailed)

		if !ok {
			t.Errorf("Error should be a *ErrTrackFailed: %v", err)
			return
		}

		if terr.Message != "some error" {
			t.Errorf("Wrong body carried in the *ErrTrackFailed: %q", terr.Message)
		}
	}

	client = New("e3bc4100330c35722740fb8c6f5abddc", ts.URL)

	assertErrTrackFailed(client.Update("1", &Update{}))
	assertErrTrackFailed(client.Track("1", "name", &Event{}))
	assertErrTrackFailed(client.Import("1", "name", &Event{}))
}
