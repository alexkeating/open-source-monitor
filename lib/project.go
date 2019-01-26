package lib

import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "errors"
import "io/ioutil"

type LatestVersion struct {
	TagName     string
	PublishedAt time.Time
	HtmlUrl     string
}

type Project interface {
	GetLatestVersion() (LatestVersion, error)
}

type Github struct {
	ProjectUrl string
	Owner      string
	Repo       string
}

type GithubLatestTag struct {
	Tag_Name     string
	Published_At string
	Html_Url     string
}

// Get the lastest version toot a message
// GET /repos/:owner/:repo/releases/latest
func (g Github) GetLatestVersion() (LatestVersion, error) {
	var bodyJson GithubLatestTag
	url := fmt.Sprintf("https://api.github.com/repos/%v/%v/releases/latest", g.Owner, g.Repo)
	resp, err := http.Get(url)
	if err != nil {
		return LatestVersion{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return LatestVersion{}, errors.New(fmt.Sprintf("Response from %v returned had a status: %v", url, resp.Status))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return LatestVersion{}, err
	}
	err = json.Unmarshal(body, &bodyJson)
	if err != nil {
		return LatestVersion{}, err
	}
	published_date, err := time.Parse(time.RFC3339, bodyJson.Published_At)
	if err != nil {
		return LatestVersion{}, err
	}
	return LatestVersion{TagName: bodyJson.Tag_Name,
		PublishedAt: published_date,
		HtmlUrl:     bodyJson.Html_Url}, nil
}
