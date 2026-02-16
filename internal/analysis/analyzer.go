package analysis

import (
	"fmt"
	"paper-scraper/internal/model"
	"regexp"
	"sort"
	"strings"
)

type DailySummary struct {
	Date          string              `json:"date"`
	TotalPapers   int                 `json:"total_papers"`
	TopTopics     []TopicCount        `json:"top_topics"`
	Breakthroughs []PaperWithOneLiner `json:"breakthroughs"`
	MajorTrends   []string            `json:"major_trends"`
}

type PaperWithOneLiner struct {
	model.Paper
	OneLiner string `json:"one_liner"`
}

type TopicCount struct {
	Topic string `json:"topic"`
	Count int    `json:"count"`
}

// 需要高亮的重点机构（简化列表）
var TopInstitutions = []string{
	"MIT", "Stanford", "Berkeley", "Carnegie Mellon", "CMU",
	"Google", "DeepMind", "Meta", "Facebook", "Microsoft",
	"Tsinghua", "Peking", "ETH", "Oxford", "Cambridge",
}

// 指示潜在突破或趋势的关键字
// 将关键字映射到显示名称
var TrendKeywords = map[string]string{
	"LLM":                    "Large Language Model",
	"Large Language Model":   "Large Language Model",
	"Generative":             "Generative AI",
	"Diffusion":              "Diffusion Models",
	"Transformer":            "Transformer",
	"Reinforcement Learning": "Reinforcement Learning",
	"Agent":                  "AI Agents",
	"Zero-shot":              "Zero-shot Learning",
	"Few-shot":               "Few-shot Learning",
	"Multimodal":             "Multimodal Learning",
	"Vision-Language":        "Vision-Language Models",
	"Graph Neural Network":   "GNN",
	"Federated":              "Federated Learning",
	"Quantum":                "Quantum Computing",
	"RAG":                    "RAG",
	"Retrieval-Augmented":    "RAG",
	"Reasoning":              "Reasoning",
	"Code Generation":        "Code Generation",
}

// 从主题提取中排除的停用词
var StopWords = map[string]bool{
	"for": true, "and": true, "the": true, "with": true, "via": true,
	"of": true, "in": true, "a": true, "an": true, "using": true,
	"based": true, "to": true, "on": true, "from": true, "by": true,
	"approach": true, "method": true, "system": true, "analysis": true,
	"learning": true, "network": true, "model": true, // Too generic
	"large": true, "new": true, "novel": true, "study": true,
	"survey": true, "review": true, "performance": true,
	"proposed": true, "propose": true,
}

func AnalyzePapers(papers []model.Paper, date string) DailySummary {
	summary := DailySummary{
		Date:        date,
		TotalPapers: len(papers),
	}

	topicFreq := make(map[string]int)
	breakthroughs := make([]PaperWithOneLiner, 0)

	// 按趋势关键字对论文进行分组，以找到特定的代表性论文
	trendGroups := make(map[string][]model.Paper)

	for _, p := range papers {
		titleLower := strings.ToLower(p.Title)
		abstractLower := strings.ToLower(p.Abstract)
		combined := titleLower + " " + abstractLower

		// 1. 主题提取（基于标题的简单词频）
		words := strings.Fields(titleLower)
		for _, w := range words {
			// 清除标点符号
			w = strings.Trim(w, ":,.-()[]\"'")
			if len(w) < 3 {
				continue
			}
			if !StopWords[w] {
				topicFreq[w]++
			}
		}

		// 同时检查特定的趋势关键字
		for kw, display := range TrendKeywords {
			if strings.Contains(combined, strings.ToLower(kw)) {
				trendGroups[display] = append(trendGroups[display], p)
				// 也将这些已知概念计为主题（权重更高？）
				topicFreq[display]++
			}
		}

		// 2. 识别“突破”/亮点
		isHighlight := false

		// 检查标题或摘要中的“强词”
		if strings.Contains(combined, "state-of-the-art") ||
			strings.Contains(combined, "outperform") ||
			strings.Contains(combined, "surpass") ||
			strings.Contains(combined, "novel") ||
			strings.Contains(combined, "first time") {
			isHighlight = true
		}

		if p.CCFClass == "A" {
			isHighlight = true
		}

		if isHighlight {
			// 提取一句话简介
			oneLiner := extractOneLiner(p.Abstract)
			breakthroughs = append(breakthroughs, PaperWithOneLiner{
				Paper:    p,
				OneLiner: oneLiner,
			})
		}
	}

	// 排序并挑选热门主题
	topics := make([]TopicCount, 0)
	for k, v := range topicFreq {
		if v > 1 { // 阈值
			topics = append(topics, TopicCount{Topic: k, Count: v})
		}
	}
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Count > topics[j].Count
	})
	if len(topics) > 10 {
		topics = topics[:10]
	}
	summary.TopTopics = topics

	// 限制突破数量
	// 优先 CCF A 或高引用（如果有）或仅按列表顺序
	sort.Slice(breakthroughs, func(i, j int) bool {
		// 优先考虑 CCF A
		if breakthroughs[i].CCFClass == "A" && breakthroughs[j].CCFClass != "A" {
			return true
		}
		if breakthroughs[i].CCFClass != "A" && breakthroughs[j].CCFClass == "A" {
			return false
		}
		return false // 保持顺序
	})
	if len(breakthroughs) > 5 {
		breakthroughs = breakthroughs[:5]
	}
	summary.Breakthroughs = breakthroughs

	// 生成包含特定论文引用的主要趋势句子
	trends := make([]string, 0)

	type TrendGroup struct {
		Name   string
		Papers []model.Paper
	}
	var sortedGroups []TrendGroup
	for k, v := range trendGroups {
		if len(v) > 0 {
			sortedGroups = append(sortedGroups, TrendGroup{Name: k, Papers: v})
		}
	}
	sort.Slice(sortedGroups, func(i, j int) bool {
		return len(sortedGroups[i].Papers) > len(sortedGroups[j].Papers)
	})

	// 挑选前 3 个趋势
	maxTrends := 3
	if len(sortedGroups) < maxTrends {
		maxTrends = len(sortedGroups)
	}

	for i := 0; i < maxTrends; i++ {
		group := sortedGroups[i]
		count := len(group.Papers)

		// 挑选一篇代表性论文（例如，摘要最长或包含 "new"）
		repPaper := group.Papers[0]
		// 简单启发式：挑选标题中包含 "propose" 或 "novel" 的论文，否则选第一篇
		for _, p := range group.Papers {
			t := strings.ToLower(p.Title)
			if strings.Contains(t, "novel") || strings.Contains(t, "new framework") {
				repPaper = p
				break
			}
		}

		// 构建叙述
		// 例如 "<b>Large Language Model</b>: 5 篇相关论文。其中 <i>'Title'</i> 提出了..."
		oneLiner := extractOneLiner(repPaper.Abstract)
		// 应要求移除激进的截断以提供更多细节
		// if len(oneLiner) > 50 {
		// 	oneLiner = oneLiner[:50] + "..."
		// }

		narrative := fmt.Sprintf("<b>%s</b>: 今日有 %d 篇相关论文。重点关注 <i>%s</i>，该研究%s",
			group.Name, count, repPaper.Title, oneLiner)

		trends = append(trends, narrative)
	}

	summary.MajorTrends = trends

	return summary
}

// extractOneLiner 尝试找到“贡献”部分并返回重要片段
func extractOneLiner(abstract string) string {
	// 1. 尝试找到开始描述贡献的关键短语
	re := regexp.MustCompile(`(?i)(we propose|we present|we introduce|this paper presents|this work|in this paper)`)
	loc := re.FindStringIndex(abstract)

	var content string
	if loc != nil {
		// 找到关键短语，提取其后的所有内容
		content = abstract[loc[0]:]
	} else {
		// 后备方案：使用整个摘要
		content = abstract
	}

	content = strings.TrimSpace(content)

	// 2. 限制长度但保持宽松（例如 600 字符）以确保上下文
	// 前端将处理展开/折叠
	const MaxLen = 600
	if len(content) > MaxLen {
		// 尝试在 MaxLen 之前的最后一个空格处截断，以避免拆分单词
		cutIndex := strings.LastIndex(content[:MaxLen], " ")
		if cutIndex > MaxLen/2 { // 确保不会截断得太早
			content = content[:cutIndex] + "..."
		} else {
			content = content[:MaxLen] + "..."
		}
	}

	return content
}
