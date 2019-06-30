package hn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	m       sync.Mutex
	stories []Story
}

var apiURL = "https://hacker-news.firebaseio.com/v0"
var maxNumberOfStories = 30

type Story struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Url    string `json:"url"`
	Score  int    `json:"score"`
	Type   string `json:"type"`
	Domain string
}

func (hnc Client) getTopStoriesIDs(numberOfStories int) ([]int, error) {
	var topStoriesIDs []int
	tsBody, err := getHTTPBody(apiURL + "/topstories.json")
	if err != nil {
		return topStoriesIDs, err
	}
	if err := json.Unmarshal(tsBody, &topStoriesIDs); err != nil {
		return nil, err
	}
	return topStoriesIDs[:numberOfStories], nil
}

func (hnc Client) getStory(id string) (Story, error) {
	var s Story
	storyBody, err := getHTTPBody(apiURL + "/item/" + id + ".json")
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(storyBody, &s); err != nil {
		log.Error(err)
	}
	if s.Url == "" {
		return s, fmt.Errorf("No URL for story %v", string(storyBody))
	} else {
		log.Info(s.Url)
	}
	domain, err := parseDomain(s.Url)
	if err != nil {
		log.Error("could not parse domain from URL:", s.Url)
	}
	s.Domain = domain
	return s, nil
}

func (hnc Client) TopStories() ([]Story, error) {
	var topStoriesIDs []int
	topStoriesIDs, err := hnc.getTopStoriesIDs(maxNumberOfStories)
	if err != nil {
		return []Story{}, err
	}
	var wg sync.WaitGroup
	wg.Add(maxNumberOfStories)

	for _, id := range topStoriesIDs {
		go func(i int) {
			defer wg.Done()
			itemID := strconv.Itoa(i)
			story, err := hnc.getStory(itemID)
			if err != nil {
				log.Error("Could not fetch story ID ", itemID, ": ", err)
				return
			}
			hnc.m.Lock()
			hnc.stories = append(hnc.stories, story)
			hnc.m.Unlock()

		}(id)
	}
	wg.Wait()

	hnc.stories = sortStories(hnc.stories)
	return hnc.stories, nil
}

func getHTTPBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, nil

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
