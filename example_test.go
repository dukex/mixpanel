package mixpanel

import "time"

func ExampleNew() {
	New("mytoken", "")
}

func ExampleNewWithSecret() {
	NewWithSecret("mytoken", "myapisecret", "")
}

func ExampleMixpanel() {
	client := New("mytoken", "")

	client.Track("1", "Sign Up", &Event{
		Properties: map[string]interface{}{
			"from": "email",
		},
	})
}

func ExamplePeople() {
	client := NewWithSecret("mytoken", "myapisecret", "")

	client.UpdateUser("1", &Update{
		Operation: "$set",
		Properties: map[string]interface{}{
			"$email":       "user@email.com",
			"$last_login":  time.Now(),
			"$created":     time.Now().String(),
			"custom_field": "cool!",
		},
	})

	client.Track("1", "Sign Up", &Event{
		Properties: map[string]interface{}{
			"from": "email",
		},
	})

	importTimestamp := time.Now().Add(-5 * 24 * time.Hour)
	client.Import("1", "Sign Up", &Event{
		Timestamp: &importTimestamp,
		Properties: map[string]interface{}{
			"subject": "topic",
		},
	})
}
