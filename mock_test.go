package mixpanel

import "fmt"

var fullfillsInterface Mixpanel = &Mock{}

func ExampleMock() {
	var people People
	client := NewMock()

	people = client.Identify("1")
	people.Update("$set", map[string]interface{}{
		"custom_field": "cool!",
	})

	people.Track("Sign Up", map[string]interface{}{
		"from": "email",
	})

	fmt.Println(client)

	// Output:
	// 1:
	//   properties:
	//     custom_field: cool!
	//   events:
	//     Sign Up:
	//       from: email
}
