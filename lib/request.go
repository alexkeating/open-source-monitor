package lib

import "net/http"
import "net/http/httputil"
import "os"
import "bytes"
import "errors"
import "fmt"
import "io"
import "encoding/json"

import log "github.com/Sirupsen/logrus"

func GetMastodonAccessToken() string {
	return os.Getenv("ACCESS_TOKEN")
}

func MastodonRequest(method, url string, body io.Reader) (*http.Response, error) {
	// I think this unction is doing too much
	// It should only create a request and add the appropriate header
	// It should be up to the calling code to execute the request
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf(`Bearer %v`, os.Getenv("ACCESS_TOKEN")))
	req.Header.Add("Content-Type", "application/json")
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"request": fmt.Sprintf("%q", dump),
	}).Debug("Request to Mastodon:")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// send status request if the update was in
type MastodonStatusRequest struct {
	Status     string `json:"status"`
	Visibility string `json:visibility`
}

func PostMastodonStatus(status string) error {
	request := MastodonStatusRequest{
		Status:     status,
		Visibility: "public",
	}
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	resp, err := MastodonRequest("POST", "https://mastodon.alexkeating.me/api/v1/statuses", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return errors.New(fmt.Sprintf("Response to %v returned had a status: %v", "mastodon.alexkeating.me", resp.Status))
	}
	return nil

}
