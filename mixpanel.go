package mixpanel

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// The Mixapanel struct store the mixpanel endpoint and the project token
type Mixpanel interface {
	// Track create a events to current distinct id
	Track(distinctId string, eventName string, properties map[string]interface{}) (*http.Response, error)

	// Identify call mixpanel 'engage' and returns People instance
	Identify(id string) People
}

// The Mixapanel struct store the mixpanel endpoint and the project token
type mixpanel struct {
	Token  string
	ApiURL string
}

// People represents a consumer, and is used on People Analytics
type People interface {
	// Track create a events to current people
	Track(eventName string, properties map[string]interface{}) (*http.Response, error)

	// Create a Update Operation to current people, see
	// https://mixpanel.com/help/reference/http
	Update(operation string, updateParams map[string]interface{}) (*http.Response, error)
}

// People represents a consumer, and is used on People Analytics
type people struct {
	m  *mixpanel
	id string
}

type trackParams struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
}

// Track create a events to current distinct id
func (m *mixpanel) Track(distinctId string, eventName string,
	properties map[string]interface{}) (*http.Response, error) {
	params := trackParams{Event: eventName}

	params.Properties = make(map[string]interface{}, 0)
	params.Properties["token"] = m.Token
	params.Properties["distinct_id"] = distinctId

	for key, value := range properties {
		params.Properties[key] = value
	}

	return m.send("track", params)
}

// Identify call mixpanel 'engage' and returns People instance
func (m *mixpanel) Identify(id string) People {
	params := map[string]interface{}{"$token": m.Token, "$distinct_id": id}
	m.send("engage", params)
	return &people{m: m, id: id}
}

// Track create a events to current people
func (p *people) Track(eventName string, properties map[string]interface{}) (*http.Response, error) {
	return p.m.Track(p.id, eventName, properties)
}

// Create a Update Operation to current people, see
// https://mixpanel.com/help/reference/http
func (p *people) Update(operation string, updateParams map[string]interface{}) (*http.Response, error) {
	params := map[string]interface{}{
		"$token":       p.m.Token,
		"$distinct_id": p.id,
	}
	params[operation] = updateParams
	return p.m.send("engage", params)
}

func (m *mixpanel) to64(data string) string {
	bytes := []byte(data)
	return base64.StdEncoding.EncodeToString(bytes)
}

func (m *mixpanel) send(eventType string, params interface{}) (*http.Response, error) {
	dataJSON, _ := json.Marshal(params)
	data := string(dataJSON)

	url := m.ApiURL + "/" + eventType + "?data=" + m.to64(data)
	return http.Get(url)
}

// New returns the client instance. If apiURL is blank, the default will be used
// ("https://api.mixpanel.com").
func New(token, apiURL string) Mixpanel {
	if apiURL == "" {
		apiURL = "https://api.mixpanel.com"
	}

	return &mixpanel{
		Token:  token,
		ApiURL: apiURL,
	}
}
