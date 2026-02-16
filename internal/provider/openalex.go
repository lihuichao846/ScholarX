package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"paper-scraper/internal/model"
)

func FetchOpenAlex(query string, limit, offset int, startDate, endDate string, sortOrder string) ([]model.Paper, error) {
	filters := []string{"concepts.id:C41008148"} // 计算机科学
	if startDate != "" && endDate != "" {
		filters = append(filters, fmt.Sprintf("from_publication_date:%s", startDate))
		filters = append(filters, fmt.Sprintf("to_publication_date:%s", endDate))
	}

	// OpenAlex 使用从 1 开始的页码
	page := (offset / limit) + 1

	params := url.Values{}
	params.Set("per-page", strconv.Itoa(limit))
	params.Set("page", strconv.Itoa(page))
	params.Set("filter", strings.Join(filters, ","))

	if sortOrder == "published_asc" {
		params.Set("sort", "publication_date:asc")
	} else {
		params.Set("sort", "publication_date:desc")
	}

	if query != "" {
		params.Set("search", query)
	}

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://api.openalex.org/works?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var oaResp model.OAResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaResp); err != nil {
		return nil, err
	}

	var papers []model.Paper
	for _, item := range oaResp.Results {
		var authors []string
		for _, ship := range item.Authorships {
			if ship.Author.DisplayName != "" {
				authors = append(authors, ship.Author.DisplayName)
			}
		}
		var categories []string
		for i, c := range item.Concepts {
			if i >= 5 {
				break
			}
			categories = append(categories, c.DisplayName)
		}

		venue := item.PrimaryLocation.Source.DisplayName
		if venue == "" {
			venue = "OpenAlex"
		}

		y := item.PublicationYear

		paper := model.Paper{
			ID:          item.ID,
			Title:       item.DisplayName,
			Authors:     authors,
			Venue:       venue,
			Year:        &y,
			Abstract:    parseOpenAlexAbstract(item.AbstractInverted),
			URL:         item.PrimaryLocation.LandingPageURL,
			Source:      "openalex",
			Categories:  categories,
			PublishedAt: item.PublicationDate,
			Citations:   item.CitedByCount,
			CCFClass:    GetCCFClass(venue),
		}
		if paper.URL == "" {
			paper.URL = item.ID
		}
		papers = append(papers, paper)
	}
	return papers, nil
}
