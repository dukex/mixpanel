mixpanel
========

Mixpanel Go Client

## Usage

Import

``` go
import "github.com/dukex/mixpanel"
```
--

Config

``` go
mixpanelToken :=   "e3bc4100330c35722740fb8c6f5abddc"
client := NewMixpanel(mixpanelToken)
```
--

Track

``` go
res, err := client.Track("13793", "Signed Up", map[string]interface{}{
    "Referred By": "Friend",
})
```
--

Identify

``` go
people := client.Identify("13793")

res, err := people.Track(map[string]interface{}{
  "Buy": "133"
})

res, err := people.Update("$set", map[string]interface{}{
  "Address":  "1313 Mockingbird Lane",
  "Birthday": "1948-01-01",
  })
```
