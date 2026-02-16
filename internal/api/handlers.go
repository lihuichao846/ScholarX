package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"paper-scraper/internal/analysis"
	"paper-scraper/internal/model"
	"paper-scraper/internal/pkg/translator"
	"paper-scraper/internal/provider"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDailySummary(c *gin.Context) {
	// 1. 确定“今天”（UTC 或本地时间）
	now := time.Now()
	// ArXiv 更新通常发生在夜间。为了确保“每日”摘要有内容，我们查看过去 48 小时的数据
	// 或者如果需要严格性，可以专门针对“今天”。
	// 在此演示中，我们将获取过去 2 天的数据以确保有内容展示。
	// 格式：YYYY-MM-DD
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// 2. 从提供商获取数据
	// 我们将搜索广泛的类别以获得“前沿”概览。
	// CS.AI, CS.CL (计算与语言), CS.CV (计算机视觉)
	// queries := []string{"cat:cs.AI", "cat:cs.CL", "cat:cs.CV"} // 暂时未使用，因为我们使用广泛搜索

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allPapers []model.Paper

	// 获取 ArXiv 数据（过去 2 天）
	// 我们会获取多一点数据以确保有足够的分析样本
	limit := 100

	wg.Add(1)
	go func() {
		defer wg.Done()
		// 使用 "cs" 类别广泛搜索或特定子类别
		// 我们将搜索 "Artificial Intelligence" 以确保覆盖面广且安全
		// 每日摘要总是需要最新的论文，因此 sortOrder 为 "published_desc"（如果未更改签名，则为提供商默认值）
		// 但我们更改了签名，所以需要传递它。
		papers, err := provider.FetchArxiv("Artificial Intelligence", limit, 0, yesterday, today, "published_desc")
		if err == nil {
			mu.Lock()
			allPapers = append(allPapers, papers...)
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// OpenAlex
		papers, err := provider.FetchOpenAlex("Artificial Intelligence", limit, 0, yesterday, today, "published_desc")
		if err == nil {
			mu.Lock()
			allPapers = append(allPapers, papers...)
			mu.Unlock()
		}
	}()

	wg.Wait()

	// 3. 分析
	summary := analysis.AnalyzePapers(allPapers, today)

	c.JSON(http.StatusOK, summary)
}

func SearchPapers(c *gin.Context) {
	query := c.Query("query")
	sourcesStr := c.DefaultQuery("sources", "arxiv,openalex")
	month := c.Query("month")
	isTopTier := c.Query("top_tier") == "true"
	ccfFilter := c.Query("ccf_level") // A, B, C, or empty
	sortOrder := c.DefaultQuery("sort", "published_desc")

	// 翻译逻辑
	searchQuery := query
	translation := ""
	if translator.ContainsChinese(query) {
		if trans, ok := translator.TranslateQuery(query); ok {
			searchQuery = trans
			translation = trans
		}
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	sources := strings.Split(sourcesStr, ",")
	sourcesSet := make(map[string]bool)
	for _, s := range sources {
		sourcesSet[s] = true
	}

	// 日期范围
	startDate, endDate := "", ""
	if month != "" {
		startDate, endDate = provider.GetMonthDateRange(month)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allPapers []model.Paper

	// 1. 获取数据

	if sourcesSet["arxiv"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			papers, err := provider.FetchArxiv(searchQuery, limit, offset, startDate, endDate, sortOrder)
			if err == nil {
				mu.Lock()
				allPapers = append(allPapers, papers...)
				mu.Unlock()
			} else {
				fmt.Println("ArXiv error:", err)
			}
		}()
	}

	if sourcesSet["openalex"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			papers, err := provider.FetchOpenAlex(searchQuery, limit, offset, startDate, endDate, sortOrder)
			if err == nil {
				mu.Lock()
				allPapers = append(allPapers, papers...)
				mu.Unlock()
			} else {
				fmt.Println("OpenAlex error:", err)
			}
		}()
	}

	wg.Wait()

	// 1.5 获取 ArXiv 论文的引用
	// 引用量数据不可捕获，故去除相关逻辑

	// 2. 过滤与排序
	var filtered []model.Paper

	// 日期解析以进行严格过滤
	var startT, endT time.Time
	if startDate != "" && endDate != "" {
		startT, _ = time.Parse("2006-01-02", startDate)
		endT, _ = time.Parse("2006-01-02", endDate)
		// 将 endT 设置为当天的结束时间
		endT = endT.Add(24*time.Hour - time.Nanosecond)
	}

	for _, p := range allPapers {
		// 严格日期过滤器（修复“月份精度”问题）
		if !startT.IsZero() && !endT.IsZero() {
			// PublishedAt 是字符串 ISO 或 YYYY-MM-DD
			pubT := provider.ParseDate(p.PublishedAt)
			if !pubT.IsZero() {
				if pubT.Before(startT) || pubT.After(endT) {
					continue
				}
			}
		}

		// 顶会/顶刊过滤器
		if isTopTier {
			venueLower := strings.ToLower(p.Venue)
			isTop := false
			for k := range provider.TopTierVenues {
				if strings.Contains(venueLower, k) {
					isTop = true
					break
				}
			}
			if !isTop {
				continue
			}
		}

		// CCF 过滤器
		if ccfFilter != "" {
			if p.CCFClass != ccfFilter {
				continue
			}
		}

		filtered = append(filtered, p)
	}

	// 按日期降序排序
	sort.Slice(filtered, func(i, j int) bool {
		if sortOrder == "published_asc" {
			return filtered[i].PublishedAt < filtered[j].PublishedAt
		}
		// 引用量排序已移除
		return filtered[i].PublishedAt > filtered[j].PublishedAt
	})

	c.JSON(http.StatusOK, model.PaperResponse{
		Count:       len(filtered),
		Items:       filtered,
		Translation: translation,
	})
}
