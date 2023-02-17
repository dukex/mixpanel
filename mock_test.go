package mixpanel

import (
	"context"
	"fmt"
	"time"
)

var fullfillsInterface Mixpanel = &Mock{}

func ExampleMock() {
	client := NewMock()

	t, _ := time.Parse(time.RFC3339, "2016-03-03T15:17:53+01:00")

	client.Update(context.TODO(), "1", &Update{
		Operation: "$set",
		Timestamp: &t,
		IP:        "127.0.0.1",
		Properties: map[string]interface{}{
			"custom_field": "cool!",
		},
	})

	client.Track(context.TODO(), "1", "Sign Up", &Event{
		IP: "1.2.3.4",
		Properties: map[string]interface{}{
			"from": "email",
		},
	})

	client.Import(context.TODO(), "1", "Sign Up", &Event{
		IP:        "1.2.3.4",
		Timestamp: &t,
		Properties: map[string]interface{}{
			"imported": true,
		},
	})

	fmt.Println(client)

	// Output:
	// 1:
	//   ip: 127.0.0.1
	//   time: 2016-03-03T15:17:53+01:00
	//   properties:
	//     custom_field: cool!
	//   events:
	//     Sign Up:
	//       IP: 1.2.3.4
	//       Timestamp:
	//       from: email
	//     Sign Up:
	//       IP: 1.2.3.4
	//       Timestamp: 2016-03-03T15:17:53+01:00
	//       imported: true

}
