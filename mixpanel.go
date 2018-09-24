package mixpanel

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var IgnoreTime *time.Time = &time.Time{}

type ErrTrackFailed struct {
	Body string
	Resp *http.Response
}

func (err *ErrTrackFailed) Error() string {
	return fmt.Sprintf("Mixpanel did not return 1 when tracking: %s", err.Body)
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type Mixpanel interface {
	// Create a mixpanel event
	Track(distinctId, eventName string, e *Event) error

	// Set properties for a mixpanel user.
	Update(distinctId string, u *Update) error

	Alias(distinctId, newId string) error
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type mixpanel struct {
	Client *http.Client
	Token  string
	ApiURL string
}

// A mixpanel event
type Event struct {
	// IP-address of the user. Leave empty to use autodetect, or set to "0" to
	// not specify an ip-address.
	IP string

	// Timestamp. Set to nil to use the current time.
	Timestamp *time.Time

	// Custom properties. At least one must be specified.
	Properties map[string]interface{}
}

// An update of a user in mixpanel
type Update struct {
	// IP-address of the user. Leave empty to use autodetect, or set to "0" to
	// not specify an ip-address at all.
	IP string

	// Timestamp. Set to nil to use the current time, or IgnoreTime to not use a
	// timestamp.
	Timestamp *time.Time

	// Update operation such as "$set", "$update" etc.
	Operation string

	// Custom properties. At least one must be specified.
	Properties map[string]interface{}
}

// Track create a events to current distinct id
func (m *mixpanel) Alias(distinctId, newId string) error {
	props := map[string]interface{}{
		"token":       m.Token,
		"distinct_id": distinctId,
		"alias":       newId,
	}

	params := map[string]interface{}{
		"event":      "$create_alias",
		"properties": props,
	}

	return m.send("track", params, false)
}

// Track create a events to current distinct id
func (m *mixpanel) Track(distinctId, eventName string, e *Event) error {
	props := map[string]interface{}{
		"token":       m.Token,
		"distinct_id": distinctId,
	}
	if e.IP != "" {
		props["ip"] = e.IP
	}
	if e.Timestamp != nil {
		props["time"] = e.Timestamp.Unix()
	}

	for key, value := range e.Properties {
		props[key] = value
	}

	params := map[string]interface{}{
		"event":      eventName,
		"properties": props,
	}

	autoGeolocate := e.IP == ""

	return m.send("track", params, autoGeolocate)
}

// Updates a user in mixpanel. See
// https://mixpanel.com/help/reference/http#people-analytics-updates
func (m *mixpanel) Update(distinctId string, u *Update) error {
	params := map[string]interface{}{
		"$token":       m.Token,
		"$distinct_id": distinctId,
	}

	if u.IP != "" {
		params["$ip"] = u.IP
	}
	if u.Timestamp == IgnoreTime {
		params["$ignore_time"] = true
	} else if u.Timestamp != nil {
		params["$time"] = u.Timestamp.Unix()
	}

	params[u.Operation] = u.Properties

	autoGeolocate := u.IP == ""

	return m.send("engage", params, autoGeolocate)
}

func (m *mixpanel) to64(data string) string {
	bytes := []byte(data)
	return base64.StdEncoding.EncodeToString(bytes)
}

func (m *mixpanel) send(eventType string, params interface{}, autoGeolocate bool) error {
	dataJSON, _ := json.Marshal(params)
	data := string(dataJSON)

	url := m.ApiURL + "/" + eventType + "?data=" + m.to64(data)

	if autoGeolocate {
		url += "&ip=1"
	}

	resp, err := m.Client.Get(url)

	if err != nil {
		return fmt.Errorf("mixpanel: %s", err.Error())
	}

	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)

	if bodyErr != nil {
		return fmt.Errorf("mixpanel: %s", bodyErr.Error())
	}

	if strBody := string(body); strBody != "1" && strBody != "1\n" {
		return &ErrTrackFailed{Body: strBody, Resp: resp}
	}

	return nil
}

// New returns the client instance. If apiURL is blank, the default will be used
// ("https://api.mixpanel.com").
func New(token, apiURL string) Mixpanel {
	return NewFromClient(http.DefaultClient, token, apiURL)
}

// Creates a client instance using the specified client instance. This is useful
// when using a proxy.
func NewFromClient(c *http.Client, token, apiURL string) Mixpanel {
	if apiURL == "" {
		apiURL = "https://api.mixpanel.com"
	}

	return &mixpanel{
		Client: c,
		Token:  token,
		ApiURL: apiURL,
	}
}
