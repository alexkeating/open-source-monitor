package main

import "time"
import "errors"
import "net/http"
import "log"
import "fmt"
import "io/ioutil"

import "encoding/json"

type Repo struct {
	Link  string
	Owner string
	Repo  string
	Type  string
}

func main() {
	var repos []Repo
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
		fmt.Println(latestVersion)
		PostMastodonStatus()
	}
	fmt.Printf("Repos: %v", repos)
	fmt.Printf("hello, world\n")
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

func PostMastodonStatus() error {
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
