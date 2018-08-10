package main

import "net/http/httputil"
import "os"
import "bytes"
import "time"
import "errors"
import "net/http"
import "fmt"
import "io/ioutil"
import "io"
import "encoding/json"

import "github.com/joho/godotenv"
import log "github.com/Sirupsen/logrus"

type Repo struct {
	Link  string
	Owner string
	Repo  string
	Type  string
}

func main() {
	var repos []Repo
	log.SetLevel(log.DebugLevel)
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	contents, err := OpenFile("projects.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(contents, &repos)
	if err != nil {
		log.Fatal(err)
	}
	for _, project := range repos {
		projectType, err := BuildProjectType(project)
		if err != nil {
			log.Fatal(err)
		}
		latestVersion, err := projectType.GetLatestVersion()
		if err != nil {
			log.Fatal(err)
		}
		lastCheck := time.Now().UTC().Sub(latestVersion.PublishedAt)
		if lastCheck.Hours() >= 24 {
			continue
		}
		year, month, day := latestVersion.PublishedAt.Date()
		log.WithFields(log.Fields{"TagName": latestVersion.TagName,
			"PublishedAt": latestVersion.PublishedAt,
			"HTMLURl":     latestVersion.HtmlUrl}).Debug("Project metadata for message:")
		PostMastodonStatus(fmt.Sprintf("Version %v of %v was published on %v! Check out the changes at the link below:\n\n%v", latestVersion.TagName, project.Repo, fmt.Sprintf("%v %vth %v", month, day, year), latestVersion.HtmlUrl))
	}
}

func BuildProjectType(project Repo) (Project, error) {
	if project.Type == "Github" {
		return Github{ProjectUrl: project.Link,
			Repo:  project.Repo,
			Owner: project.Owner}, nil
	}
	return nil, errors.New(fmt.Sprintf("Methods not implmented for project type %v", project.Type))
}

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
	if resp.StatusCode > 399 {
		return LatestVersion{}, errors.New(fmt.Sprintf("Response from %v returned had a status: %v", url, resp.Status))
	}
	if err != nil {
		return LatestVersion{}, err
	}
	defer resp.Body.Close()
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

// Get mastodon access token
// Then Post status
// Authorize github account
// Testing documentation
// Clean up split things into seperate packages
// Logging

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

func OpenFile(path string) ([]byte, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// The goal of this bot is to keep me up to date on when open source projects upgrade and potential contribution opportunities

// Have a list of of repos in yaml format
// go through each repo check if the version has been updated
// If the version has beem updated send a toot with a link to the change log possible more info
// Get all issues find the hottest issues or ones tagged with first time contributor or somehting similar
