package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type S2Paper struct {
	PaperID       string `json:"paperId"`
	CitationCount int    `json:"citationCount"`
}

func FetchCitations(arxivIDs []string) (map[string]int, error) {
	if len(arxivIDs) == 0 {
		return nil, nil
	}

	// 准备 ID
	// S2 期望格式为 "ARXIV:ID"。
	// 我们的 ID 通常是像 "http://arxiv.org/abs/2106.12345v1" 这样的 URL
	// 我们需要提取 ID 部分。

	payloadIDs := make([]string, 0, len(arxivIDs))
	// 从 S2 请求 ID 映射到原始 ID（或者我们可以直接匹配剥离后的 ID）
	// 实际上，S2 返回的是 "paperId"，这是它的内部 ID (SHA)，而不是我们发送的 arXiv ID。
	// 等等。响应是对应输入列表顺序的对象列表吗？还是 null？
	// 批量端点返回一个对象数组。
	// 文档说明：“返回论文对象列表，顺序与输入 ID 相同。”
	// 如果未找到论文，该条目为 null。

	for _, rawID := range arxivIDs {
		// 清洗 ID
		cleanID := rawID
		cleanID = strings.TrimPrefix(cleanID, "http://arxiv.org/abs/")
		cleanID = strings.TrimPrefix(cleanID, "https://arxiv.org/abs/")
		// 移除版本后缀（例如 v1, v2）以获得更好的 S2 匹配
		if idx := strings.LastIndex(cleanID, "v"); idx > 0 && idx < len(cleanID)-1 {
			// 检查后续字符是否为数字
			isVersion := true
			for _, r := range cleanID[idx+1:] {
				if r < '0' || r > '9' {
					isVersion = false
					break
				}
			}
			if isVersion {
				cleanID = cleanID[:idx]
			}
		}
		// 通常 ARXIV:ID 是有效的。
		payloadIDs = append(payloadIDs, "ARXIV:"+cleanID)
	}

	requestBody, err := json.Marshal(map[string][]string{
		"ids": payloadIDs,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.semanticscholar.org/graph/v1/paper/batch?fields=citationCount", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("S2 status %d", resp.StatusCode)
	}

	// 注意：S2 可能会对某些条目返回 null。
	// 我们需要解码为指针切片或 raw json.RawMessage 来处理 null 吗？
	// 或者只是结构体切片，如果是 null，它可能会报错或是零值？
	// “如果未找到论文，列表中的相应条目将为 null。”
	// 所以我们需要解码为 []*S2Paper
	var rawResults []*S2Paper
	if err := json.NewDecoder(resp.Body).Decode(&rawResults); err != nil {
		return nil, err
	}

	citationMap := make(map[string]int)
	for i, res := range rawResults {
		if res != nil {
			// 映射回原始 ArXiv ID（使用索引）
			if i < len(arxivIDs) {
				citationMap[arxivIDs[i]] = res.CitationCount
			}
		}
	}

	return citationMap, nil
}
