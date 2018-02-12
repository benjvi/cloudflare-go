package cloudflare

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	rateLimitID                = "72dae2fc158942f2adb1dd2a3d4143bc"
	testZoneID                 = "abcd123"
	serverRateLimitDescription = `{
	"id": "72dae2fc158942f2adb1dd2a3d4143bc",
	"disabled": false,
	"description": "test",
	"match": {
	"request": {
	  "methods": [
		"_ALL_"
	  ],
	  "schemes": [
		"_ALL_"
	  ],
	  "url": "exampledomain.com/test-rate-limit"
	},
	"response": {
	  "origin_traffic": true
	}
	},
	"login_protect": false,
	"threshold": 50,
	"period": 1,
	"action": {
		"mode": "ban",
		"timeout": 60
	}
}
`
)

var expectedRateLimitStruct = RateLimit{
	ID:          "72dae2fc158942f2adb1dd2a3d4143bc",
	Disabled:    false,
	Description: "test",
	Match: RateLimitTrafficMatcher{
		Request: RateLimitRequestMatcher{
			Methods:    []string{"_ALL_"},
			Schemes:    []string{"_ALL_"},
			URLPattern: "exampledomain.com/test-rate-limit",
		},
		Response: RateLimitResponseMatcher{
			OriginTraffic: true,
		},
	},
	Threshold: 50,
	Period:    1,
	Action: RateLimitAction{
		Mode:    "ban",
		Timeout: 60,
	},
}
var expectedRateLimitStructUpdated = RateLimit{
	ID:          "72dae2fc158942f2adb1dd2a3d4143bc",
	Disabled:    false,
	Description: "test",
	Match: RateLimitTrafficMatcher{
		Request: RateLimitRequestMatcher{
			Methods:    []string{"_ALL_"},
			Schemes:    []string{"_ALL_"},
			URLPattern: "exampledomain.com/test-rate-limit",
		},
		Response: RateLimitResponseMatcher{
			OriginTraffic: true,
		},
	},
	Threshold: 50,
	Period:    1,
	Action: RateLimitAction{
		Mode:    "ban",
		Timeout: 60,
	},
}

func TestListRateLimits(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": [
			%s
		  ],
		  "success": true,
		  "errors": null,
		  "messages": null,
		  "result_info": {
			"page": 1,
			"per_page": 25,
			"count": 1,
			"total_count": 1
		  }
		}
		`, serverRateLimitDescription)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/rate_limits", handler)
	want := []RateLimit{expectedRateLimitStruct}

	actual, err := client.ListRateLimits(testZoneID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetRateLimit(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, serverRateLimitDescription)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/rate_limits/"+rateLimitID, handler)
	want := expectedRateLimitStruct

	actual, err := client.RateLimit(testZoneID, rateLimitID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateRateLimit(t *testing.T) {
	setup()
	defer teardown()
	newRateLimit := RateLimit{
		Description: "test",
		Match: RateLimitTrafficMatcher{
			Request: RateLimitRequestMatcher{
				URLPattern: "exampledomain.com/test-rate-limit",
			},
		},
		Period:    1,
		Threshold: 50,
		Action: RateLimitAction{
			Mode:    "ban",
			Timeout: 60,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, serverRateLimitDescription)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/rate_limits", handler)
	want := expectedRateLimitStruct

	actual, err := client.CreateRateLimit(testZoneID, newRateLimit)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateRateLimitWithZeroedThreshold(t *testing.T) {
	setup()
	defer teardown()
	newRateLimit := RateLimit{
		Description: "test",
		Match: RateLimitTrafficMatcher{
			Request: RateLimitRequestMatcher{
				URLPattern: "exampledomain.com/test-rate-limit",
			},
		},
		Period:    0, // 0 is the default values if int fields are not set
		Threshold: 0,
		Action: RateLimitAction{
			Mode:    "ban",
			Timeout: 60,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.WriteHeader(400)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": null,
		  "success": false,
		  "errors": [{ "message": "ratelimit.api.validation_error:threshold is too low and must be at least 2,sample_rate is too low and must be at least 1 second" } ],
		  "messages": null
		}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/rate_limits", handler)

	actual, err := client.CreateRateLimit(testZoneID, newRateLimit)
	assert.Error(t, err)
	assert.Equal(t, RateLimit{}, actual)
}

func TestUpdateRateLimit(t *testing.T) {
	setup()
	defer teardown()
	newRateLimit := RateLimit{
		Description: "test-2",
		Match: RateLimitTrafficMatcher{
			Request: RateLimitRequestMatcher{
				URLPattern: "exampledomain.com/test-rate-limit-2",
			},
		},
		Period:    2,
		Threshold: 100,
		Action: RateLimitAction{
			Mode:    "ban",
			Timeout: 600,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "PUT", "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, serverRateLimitDescription)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/rate_limits/"+rateLimitID, handler)
	want := expectedRateLimitStruct

	actual, err := client.UpdateRateLimit(testZoneID, rateLimitID, newRateLimit)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteRateLimit(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "DELETE", "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": null,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/rate_limits/"+rateLimitID, handler)

	err := client.DeleteRateLimit(testZoneID, rateLimitID)
	assert.NoError(t, err)
}
