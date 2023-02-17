package mixpanel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Mock Mixpanel client which can be used in unit tests.
type Mock struct {
	// All people identified, mapped by distinctId
	people sync.Map
}

func NewMock() *Mock {
	return &Mock{
		people: sync.Map{},
	}
}

func (m *Mock) String() string {
	str := ""
	m.people.Range(func(id, p interface{}) bool {
		str += id.(string) + ":\n" + p.(*MockPeople).String()
		return true
	})
	return str
}

func (m *Mock) Reset() {
	m.people = sync.Map{}
}

// People identifies a user. The user will be added to the people map.
func (m *Mock) People(distinctId string) *MockPeople {
	p, _ := m.people.LoadOrStore(distinctId, &MockPeople{
		Properties: map[string]interface{}{},
	})

	return p.(*MockPeople)
}

func (m *Mock) Track(_ context.Context, distinctId, eventName string, e *Event) error {
	p := m.People(distinctId)
	p.Events = append(p.Events, MockEvent{
		Event: *e,
		Name:  eventName,
	})
	return nil
}

func (m *Mock) Import(_ context.Context, distinctId, eventName string, e *Event) error {
	p := m.People(distinctId)
	p.Events = append(p.Events, MockEvent{
		Event: *e,
		Name:  eventName,
	})
	return nil
}

type MockPeople struct {
	Properties map[string]interface{}
	Time       *time.Time
	IP         string
	Events     []MockEvent
}

func (mp *MockPeople) String() string {
	timeStr := ""
	if mp.Time != nil {
		timeStr = mp.Time.Format(time.RFC3339)
	}

	str := fmt.Sprintf("  ip: %s\n  time: %s\n", mp.IP, timeStr)
	str += "  properties:\n"
	for key, val := range mp.Properties {
		str += fmt.Sprintf("    %s: %v\n", key, val)
	}
	str += "  events:\n"
	for _, event := range mp.Events {
		str += "    " + event.Name + ":\n"
		str += fmt.Sprintf("      IP: %s\n", event.IP)
		if event.Timestamp != nil {
			str += fmt.Sprintf(
				"      Timestamp: %s\n", event.Timestamp.Format(time.RFC3339),
			)
		} else {
			str += "      Timestamp:\n"
		}
		for key, val := range event.Properties {
			str += fmt.Sprintf("      %s: %v\n", key, val)
		}
	}
	return str
}

func (m *Mock) Update(ctx context.Context, distinctId string, u *Update) error {
	return m.UpdateUser(ctx, distinctId, u)
}

func (m *Mock) UpdateUser(_ context.Context, distinctId string, u *Update) error {
	p := m.People(distinctId)

	if u.IP != "" {
		p.IP = u.IP
	}
	if u.Timestamp != nil && u.Timestamp != IgnoreTime {
		p.Time = u.Timestamp
	}

	switch u.Operation {
	case "$set", "$set_once":
		for key, val := range u.Properties {
			p.Properties[key] = val
		}
	default:
		return errors.New("mixpanel.Mock only supports the $set and $set_once operations")
	}

	return nil
}

func (m *Mock) UnionUser(_ context.Context, userID string, u *Update) error {
	return nil
}

func (m *Mock) UpdateGroup(_ context.Context, groupKey, groupUser string, u *Update) error {
	return nil
}

func (m *Mock) UnionGroup(_ context.Context, groupKey, groupUser string, u *Update) error {
	return nil
}

func (m *Mock) Alias(_ context.Context, distinctId, newId string) error {
	return nil
}

type MockEvent struct {
	Event
	Name string
}
