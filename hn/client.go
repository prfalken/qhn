package hn

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Client struct{}

var apiURL = "https://hacker-news.firebaseio.com/v0"
var maxNumberOfStories = 30

type Story struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Url    string `json:"url"`
	Score  int    `json:"score"`
	Domain string
}

func (hnc Client) getTopStoriesIDs(numberOfStories int) []int {
	var topStoriesIDs []int
	tsBody := getHTTPBody(apiURL + "/topstories.json")
	if err := json.Unmarshal(tsBody, &topStoriesIDs); err != nil {
		log.Error(err)
	}
	return topStoriesIDs[:numberOfStories]
}

func (hnc Client) getStory(id string) Story {
	var s Story
	storyBody := getHTTPBody(apiURL + "/item/" + id + ".json")
	if err := json.Unmarshal(storyBody, &s); err != nil {
		log.Error(err)
	}
	log.Info(s.Url)
	domain, err := parseDomain(s.Url)
	if err != nil {
		log.Error("could not parse domain from URL:", s.Url)
	}
	s.Domain = domain
	return s
}

func (hnc Client) TopStories() []Story {
	var topStories []Story
	topStoriesIDs := hnc.getTopStoriesIDs(maxNumberOfStories)

	var wg sync.WaitGroup
	wg.Add(maxNumberOfStories)

	for _, id := range topStoriesIDs {
		go func(i int) {
			defer wg.Done()
			itemID := strconv.Itoa(i)
			story := hnc.getStory(itemID)
			topStories = append(topStories, story)

		}(id)
	}
	wg.Wait()

	topStories = sortStories(topStories)
	return topStories
}

func getHTTPBody(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("could not get HN api")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body

}

func sortStories(stories []Story) []Story {
	sort.Slice(stories[:], func(i, j int) bool {
		return stories[i].Score > stories[j].Score
	})
	return stories
}

func parseDomain(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}
