package mixpanel

import (
	"time"
)

func ExampleNew() {
	New("mytoken", "")
}

func ExampleMixpanel() {
	client := New("mytoken", "")

	client.Track("1", "Sign Up", map[string]interface{}{
		"from": "email",
	})
}

func ExamplePeople() {
	var people People
	client := New("mytoken", "")

	people = client.Identify("1")
	people.Update("$set", map[string]interface{}{
		"$email":       "user@email.com",
		"$last_login":  time.Now(),
		"$created":     time.Now().String(),
		"custom_field": "cool!",
	})

	people.Track("Sign Up", map[string]interface{}{
		"from": "email",
	})
}
