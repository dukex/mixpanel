package mixpanel

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type Request struct {
	*http.Request
	DecodedBody string
}

var (
	ts          *httptest.Server
	client      Mixpanel
	LastRequest *Request
)

func setup() {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("1\n"))
		body, _ := io.ReadAll(r.Body)
		LastRequest = &Request{
			r,
			decodeData(string(body)),
		}
	}))

	client = NewWithSecret("e3bc4100330c35722740fb8c6f5abddc", "mysecret", ts.URL)
}

func teardown() {
	ts.Close()
}

func decodeData(url string) string {
	data := strings.Split(url, "data=")[1]
	decoded, _ := base64.StdEncoding.DecodeString(data)
	return string(decoded[:])
}

func assert(t *testing.T, desc string, current interface{}, wanted interface{}) {
	if !reflect.DeepEqual(current, wanted) {
		t.Errorf("%s returned %+v, want %+v", desc,
			current, wanted)
	}
}

func TestTrack(t *testing.T) {
	setup()
	defer teardown()

	client.Track(context.TODO(), "13793", "Signed Up", &Event{
		Properties: map[string]interface{}{
			"Referred By": "Friend",
		},
	})

	wantedBody := "{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}"

	assert(t, "body", LastRequest.DecodedBody, wantedBody)
	assert(t, "path", LastRequest.URL.Path, "/track")
}

func TestImport(t *testing.T) {
	setup()
	defer teardown()

	importTime := time.Now().Add(-5 * 24 * time.Hour)

	client.Import(context.TODO(), "13793", "Signed Up", &Event{
		Properties: map[string]interface{}{
			"Referred By": "Friend",
		},
		Timestamp: &importTime,
	})

	wantedBody := fmt.Sprintf("{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"time\":%d,\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}", importTime.Unix())

	assert(t, "body", LastRequest.DecodedBody, wantedBody)
	assert(t, "path", LastRequest.URL.Path, "/import")
}

func TestGroupOperations(t *testing.T) {
	setup()
	defer teardown()

	client.UpdateGroup(context.TODO(), "company_id", "11", &Update{
		Operation: "$set",
		Properties: map[string]interface{}{
			"Address":  "1313 Mockingbird Lane",
			"Birthday": "1948-01-01",
		},
	})

	wantedBody := "{\"$group_id\":\"11\",\"$group_key\":\"company_id\",\"$set\":{\"Address\":\"1313 Mockingbird Lane\",\"Birthday\":\"1948-01-01\"},\"$token\":\"e3bc4100330c35722740fb8c6f5abddc\"}"

	assert(t, "body", LastRequest.DecodedBody, wantedBody)
	assert(t, "path", LastRequest.URL.Path, "/groups")
}

func TestUpdateUser(t *testing.T) {
	setup()
	defer teardown()

	client.UpdateUser(context.TODO(), "13793", &Update{
		Operation: "$set",
		Properties: map[string]interface{}{
			"Address":  "1313 Mockingbird Lane",
			"Birthday": "1948-01-01",
		},
	})

	wantedBody := "{\"$distinct_id\":\"13793\",\"$set\":{\"Address\":\"1313 Mockingbird Lane\",\"Birthday\":\"1948-01-01\"},\"$token\":\"e3bc4100330c35722740fb8c6f5abddc\"}"

	assert(t, "body", LastRequest.DecodedBody, wantedBody)
	assert(t, "path", LastRequest.URL.Path, "/engage")
}

func TestError(t *testing.T) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"error": "some error", "status": 0}`))
		LastRequest = &Request{r, ""}
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

		assert(t, "error", "error=some error; status=0; httpCode=200", terr.Message)
	}

	client = New("e3bc4100330c35722740fb8c6f5abddc", ts.URL)

	assertErrTrackFailed(client.UpdateUser(context.TODO(), "1", &Update{}))
	assertErrTrackFailed(client.Track(context.TODO(), "1", "name", &Event{}))
	assertErrTrackFailed(client.Import(context.TODO(), "1", "name", &Event{}))
}
