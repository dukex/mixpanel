package mixpanel

import (
	"errors"
	"fmt"
)

// Mocked version of Mixpanel which can be used in unit tests.
type Mock struct {
	// All People identified, mapped by distinctId
	People map[string]*MockPeople
}

func NewMock() *Mock {
	return &Mock{
		People: map[string]*MockPeople{},
	}
}

func (m *Mock) String() string {
	str := ""
	for id, p := range m.People {
		str += id + ":\n" + p.String()
	}
	return str
}

// Identifies a user. The user will be added to the People map.
func (m *Mock) Identify(distinctId string) People {
	if p, ok := m.People[distinctId]; ok {
		return p
	} else {
		p := &MockPeople{
			Properties: map[string]interface{}{},
		}
		m.People[distinctId] = p
		return p
	}
}

func (m *Mock) Track(distinctId string, eventName string, properties map[string]interface{}) error {
	return m.Identify(distinctId).Track(eventName, properties)
}

type MockPeople struct {
	Properties map[string]interface{}
	Events     []MockEvent
}

func (mp *MockPeople) String() string {
	str := ""
	str += "  properties:\n"
	for key, val := range mp.Properties {
		str += fmt.Sprintf("    %s: %v\n", key, val)
	}
	str += "  events:\n"
	for _, event := range mp.Events {
		str += "    " + event.Name + ":\n"
		for key, val := range event.Properties {
			str += fmt.Sprintf("      %s: %v\n", key, val)
		}
	}
	return str
}

func (mp *MockPeople) Track(eventName string, properties map[string]interface{}) error {
	mp.Events = append(mp.Events, MockEvent{eventName, properties})
	return nil
}

func (mp *MockPeople) Update(operation string, updateParams map[string]interface{}) error {
	switch operation {
	case "$set":
		for key, val := range updateParams {
			mp.Properties[key] = val
		}
	default:
		return errors.New("mixpanel.Mock only supports the $set operation")
	}

	return nil
}

type MockEvent struct {
	Name       string
	Properties map[string]interface{}
}
