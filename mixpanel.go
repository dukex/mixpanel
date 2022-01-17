package mixpanel

import (
	"bytes"
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
	return fmt.Sprintf("mixpanel did not return 1 when tracking: %s", err.Message)
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type Mixpanel interface {
	// Create a mixpanel event using the track api
	Track(distinctId, eventName string, e *Event) error

	// Create a mixpanel event using the import api
	Import(distinctId, eventName string, e *Event) error

	// Set properties for a mixpanel user.
	Update(distinctId string, u *Update) error

	// Create an alias for an existing distinct id
	Alias(distinctId, newId string) error

	SetProjectID(string)

	SetAuth(username string, password string)
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type mixpanel struct {
	Client    *http.Client
	Token     string
	Secret    string
	ApiURL    string
	ProjectID string
	AuthUser  string
	AuthPass  string
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

// Update will update of a user in mixpanel
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

	a := []interface{}{
		params,
	}

	return m.send("track", a, false)
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

	a := []interface{}{
		params,
	}

	return m.send("track", a, autoGeolocate)
}

// Import create an event for an existing distinct id
// See https://developer.mixpanel.com/docs/importing-old-events
func (m *mixpanel) Import(distinctId, eventName string, e *Event) error {
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

	a := []interface{}{
		params,
	}

	return m.send("import", a, autoGeolocate)
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

	a := []interface{}{
		params,
	}

	return m.send("engage", a, autoGeolocate)
}

func (m *mixpanel) to64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func (m *mixpanel) send(eventType string, params []interface{}, autoGeolocate bool) error {
	data, err := json.Marshal(params)

	if err != nil {
		return err
	}

	url := m.ApiURL + "/" + eventType + "?project_id=" + m.ProjectID

	if autoGeolocate {
		url += "&ip=1"
	}

	wrapErr := func(err error) error {
		return &MixpanelError{URL: url, Err: err}
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")

	if m.AuthUser != "" && m.AuthPass != "" {
		req.SetBasicAuth(m.AuthUser, m.AuthPass)
	} else if m.Secret != "" {
		req.SetBasicAuth(m.Secret, "")
	}

	resp, err := m.Client.Do(req)

	if err != nil {
		return wrapErr(err)
	}

	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)

	if bodyErr != nil {
		return wrapErr(bodyErr)
	}

	if string(body) != "1" {
		errMsg := fmt.Sprintf("body=%d; httpCode=%d", body, resp.StatusCode)
		return wrapErr(&ErrTrackFailed{Message: errMsg})
	}

	return nil
}

func (m *mixpanel) SetProjectID(ID string) {
	m.ProjectID = ID
}

func (m *mixpanel) SetAuth(username string, password string) {
	m.AuthUser = username
	m.AuthPass = password
}

// New returns the client instance. If apiURL is blank, the default will be used
// ("https://api.mixpanel.com").
func New(token, apiURL string) Mixpanel {
	return NewFromClient(http.DefaultClient, token, apiURL)
}

// NewWithSecret returns the client instance using a secret.If apiURL is blank,
// the default will be used ("https://api.mixpanel.com").
func NewWithSecret(token, secret, apiURL string) Mixpanel {
	return NewFromClientWithSecret(http.DefaultClient, token, secret, apiURL)
}

// NewFromClient creates a client instance using the specified client instance. This is useful
// when using a proxy.
func NewFromClient(c *http.Client, token, apiURL string) Mixpanel {
	return NewFromClientWithSecret(c, token, "", apiURL)
}

// NewFromClientWithSecret creates a client instance using the specified client instance and secret.
func NewFromClientWithSecret(c *http.Client, token, secret, apiURL string) Mixpanel {
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
