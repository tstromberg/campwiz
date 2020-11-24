package metasrc

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"k8s.io/klog/v2"
)

var (
	maxBingAge   = 90 * 24 * time.Hour
	bingEndpoint = "https://api.bing.microsoft.com/v7.0/search"
)

type BingHit struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	URL              string    `json:"url"`
	IsFamilyFriendly bool      `json:"isFamilyFriendly"`
	DisplayURL       string    `json:"displayUrl"`
	Snippet          string    `json:"snippet"`
	DateLastCrawled  time.Time `json:"dateLastCrawled"`
	SearchTags       []struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	} `json:"searchTags,omitempty"`
	About []struct {
		Name string `json:"name"`
	} `json:"about,omitempty"`
}

// This struct formats the answers provided by the Bing Web Search API.
type BingAnswer struct {
	Type         string `json:"_type"`
	QueryContext struct {
		OriginalQuery string `json:"originalQuery"`
	} `json:"queryContext"`
	WebPages struct {
		WebSearchURL          string    `json:"webSearchUrl"`
		TotalEstimatedMatches int       `json:"totalEstimatedMatches"`
		Value                 []BingHit `json:"value"`
	} `json:"webPages"`
	RelatedSearches struct {
		ID    string `json:"id"`
		Value []struct {
			Text         string `json:"text"`
			DisplayText  string `json:"displayText"`
			WebSearchURL string `json:"webSearchUrl"`
		} `json:"value"`
	} `json:"relatedSearches"`
	RankingResponse struct {
		Mainline struct {
			Items []struct {
				AnswerType  string `json:"answerType"`
				ResultIndex int    `json:"resultIndex"`
				Value       struct {
					ID string `json:"id"`
				} `json:"value"`
			} `json:"items"`
		} `json:"mainline"`
		Sidebar struct {
			Items []struct {
				AnswerType string `json:"answerType"`
				Value      struct {
					ID string `json:"id"`
				} `json:"value"`
			} `json:"items"`
		} `json:"sidebar"`
	} `json:"rankingResponse"`
}

func BingSearch(cs cache.Store, s string) (BingAnswer, error) {
	req := cache.Request{
		URL:  bingEndpoint,
		Form: url.Values{"q": {s}},
		Headers: map[string]string{
			"Ocp-Apim-Subscription-Key": os.Getenv("BING_TOKEN"),
		},
		MaxAge: maxBingAge,
	}

	var ans BingAnswer
	resp, err := cache.Fetch(req, cs)
	if err != nil {
		return ans, fmt.Errorf("cache fetch: %v", err)
	}
	klog.Infof("%s cache state: %v", resp.URL, resp.Cached)

	err = json.Unmarshal(resp.Body, &ans)
	if err != nil {
		return ans, fmt.Errorf("unmarshal: %v", err)
	}
	return ans, nil
}
