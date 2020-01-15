package mixpanel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var IgnoreTime *time.Time = &time.Time{}

type MixpanelError struct {
	URL string
	Err error
}

func (err *MixpanelError) Cause() error {
	return err.Err
}

func (err *MixpanelError) Error() string {
	return "mixpanel: " + err.Err.Error()
}

type ErrTrackFailed struct {
	Message string
}

func (err *ErrTrackFailed) Error() string {
	return fmt.Sprintf("Mixpanel did not return 1 when tracking: %s", err.Message)
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type Mixpanel interface {
	// Create a mixpanel event
	Track(distinctId, eventName string, e *Event) error

	// Set properties for a mixpanel user.
	Update(distinctId string, u *Update) error

	// Create an alias for an existing distinct id
	Alias(distinctId, newId string) error
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type mixpanel struct {
	Client *http.Client
	Token  string
	Secret string
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

// Alias create an alias for an existing distinct id
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

// Track create an event for an existing distinct id
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

// Update updates a user in mixpanel. See
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

func (m *mixpanel) to64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func (m *mixpanel) send(eventType string, params interface{}, autoGeolocate bool) error {
	data, err := json.Marshal(params)

	if err != nil {
		return err
	}

	url := m.ApiURL + "/" + eventType + "?data=" + m.to64(data) + "&verbose=1"

	if autoGeolocate {
		url += "&ip=1"
	}

	wrapErr := func(err error) error {
		return &MixpanelError{URL: url, Err: err}
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(m.Secret, "")
	resp, err := m.Client.Do(req)

	if err != nil {
		return wrapErr(err)
	}

	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)

	if bodyErr != nil {
		return wrapErr(bodyErr)
	}

	type verboseResponse struct {
		Error  string `json:"error"`
		Status int    `json:"status"`
	}

	var jsonBody verboseResponse
	json.Unmarshal(body, &jsonBody)

	if jsonBody.Status != 1 {
		return wrapErr(&ErrTrackFailed{Message: jsonBody.Error})
	}

	return nil
}

// New returns the client instance. If apiURL is blank, the default will be used
// ("https://api.mixpanel.com").
func New(token, secret, apiURL string) Mixpanel {
	return NewFromClient(http.DefaultClient, token, secret, apiURL)
}

// Creates a client instance using the specified client instance. This is useful
// when using a proxy.
func NewFromClient(c *http.Client, token, secret, apiURL string) Mixpanel {
	if apiURL == "" {
		apiURL = "https://api.mixpanel.com"
	}

	return &mixpanel{
		Client: c,
		Token:  token,
		Secret: secret,
		ApiURL: apiURL,
	}
}
