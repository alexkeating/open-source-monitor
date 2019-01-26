package lib

import "fmt"
import "testing"

func TestRequestMastodonRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		method string
		url    string
		err    string
	}{
		{name: "normal case", method: "POST", url: "https://test.com", err: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := MastodonRequest(tc.method, tc.url, nil)
			if tc.err != "" {
				fmt.Printf("Placeholder")
			}
			if err != nil {
				t.Fatalf("Could not create MastodonRequest: %v", err)
			}
			if len(req.Header.Get("Authorization")) == 0 {
				t.Errorf("Header Authorization is missing from the MastodonRequest")
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Content type is incorrect")
			}
			if req.Method != tc.method {
				t.Errorf("Method for MastodonRequest is incorrect")
			}
			if req.URL.String() != tc.url {
				t.Errorf("Url for MastodonRequest is incorrect")
			}

		})
	}

}
