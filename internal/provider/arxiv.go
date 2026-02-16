package provider

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"paper-scraper/internal/model"
)

func FetchArxiv(query string, limit, offset int, startDate, endDate string, sortOrder string) ([]model.Paper, error) {
	searchParts := []string{"cat:cs.*"}
	if startDate != "" && endDate != "" {
		start := strings.ReplaceAll(startDate, "-", "") + "0000"
		end := strings.ReplaceAll(endDate, "-", "") + "2359"
		searchParts = append(searchParts, fmt.Sprintf("submittedDate:[%s TO %s]", start, end))
	}
	if query != "" {
		searchParts = append(searchParts, fmt.Sprintf("all:%s", query))
	}

	params := url.Values{}
	params.Set("search_query", strings.Join(searchParts, "+AND+"))
	params.Set("start", strconv.Itoa(offset))
	params.Set("max_results", strconv.Itoa(limit))
	params.Set("sortBy", "submittedDate")

	if sortOrder == "published_asc" {
		params.Set("sortOrder", "ascending")
	} else {
		params.Set("sortOrder", "descending")
	}

	apiURL := "http://export.arxiv.org/api/query?" + params.Encode()
	apiURL = strings.ReplaceAll(apiURL, "%2BAND%2B", "+AND+")

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var feed model.AtomFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	var papers []model.Paper
	for _, entry := range feed.Entries {
		var authors []string
		for _, a := range entry.Authors {
			authors = append(authors, a.Name)
		}
		var categories []string
		for _, c := range entry.Category {
			if c.Term != "" {
				categories = append(categories, c.Term)
			}
		}

		venue := "arXiv"
		if entry.JournalRef != "" {
			venue = entry.JournalRef
		} else if entry.Comment != "" {
			venue = entry.Comment
		}

		year := 0
		pubDate := ParseDate(entry.Published)
		if !pubDate.IsZero() {
			year = pubDate.Year()
		}

		title := strings.TrimSpace(strings.ReplaceAll(entry.Title, "\n", " "))
		abstract := strings.TrimSpace(strings.ReplaceAll(entry.Summary, "\n", " "))

		paper := model.Paper{
			ID:          entry.ID,
			Title:       title,
			Authors:     authors,
			Venue:       venue,
			Year:        &year,
			Abstract:    abstract,
			URL:         entry.ID,
			Source:      "arxiv",
			Categories:  categories,
			PublishedAt: entry.Published,
			Citations:   0,
			CCFClass:    GetCCFClass(venue),
		}
		for _, l := range entry.Links {
			if l.Rel == "alternate" {
				paper.URL = l.Href
				break
			}
		}

		papers = append(papers, paper)
	}
	return papers, nil
}
