package mixpanel

import "time"

func ExampleNew() {
	New("mytoken", "")
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
	client := New("mytoken", "")

	client.Update("1", &Update{
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
}
