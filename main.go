package main

import "time"
import "errors"
import "fmt"
import "encoding/json"

import "github.com/joho/godotenv"
import log "github.com/Sirupsen/logrus"
import "github.com/alexkeating/open-source-monitor/lib"

type Repo struct {
	Link  string
	Owner string
	Repo  string
	Type  string
}

// Get mastodon access token
// Then Post status
// Authorize github account
// Testing documentation
// Clean up split things into seperate packages
// Logging

// The goal of this bot is to keep me up to date on when open source projects upgrade and potential contribution opportunities

// Have a list of of repos in yaml format
// go through each repo check if the version has been updated
// If the version has beem updated send a toot with a link to the change log possible more info
// Get all issues find the hottest issues or ones tagged with first time contributor or somehting similar
func main() {
	var repos []Repo
	log.SetLevel(log.DebugLevel)
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	contents, err := lib.OpenFile("projects.json")
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
		lib.PostMastodonStatus(fmt.Sprintf("Version %v of %v was published on %v! Check out the changes at the link below:\n\n%v", latestVersion.TagName, project.Repo, fmt.Sprintf("%v %vth %v", month, day, year), latestVersion.HtmlUrl))
	}
}

func BuildProjectType(project Repo) (lib.Project, error) {
	if project.Type == "Github" {
		return lib.Github{ProjectUrl: project.Link,
			Repo:  project.Repo,
			Owner: project.Owner}, nil
	}
	return nil, errors.New(fmt.Sprintf("Methods not implmented for project type %v", project.Type))
}
