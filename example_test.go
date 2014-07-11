package mixpanel

import (
	"time"
)

func ExampleNewMixpanel() {
	NewMixpanel("mytoken")
}

func ExampleMixpanel_Identify() {
	client := NewMixpanel("mytoken")

	client.Identify("1")
}

func ExampleTrack() {
	client := NewMixpanel("mytoken")

	client.Track("1", "Sign Up", map[string]interface{}{
		"from": "email",
	})
}

func ExamplePeople_Update() {
	var people *People
	client := NewMixpanel("mytoken")

	people = client.Identify("1")
	people.Update("$set", map[string]interface{}{
		"$email":       "user@email.com",
		"$last_login":  time.Now(),
		"$created":     time.Now().String(),
		"custom_field": "cool!",
	})
}

func ExamplePeople_Track() {
	var people *People
	client := NewMixpanel("mytoken")

	people = client.Identify("1")
	people.Track("Sign Up", map[string]interface{}{
		"from": "email",
	})
}
