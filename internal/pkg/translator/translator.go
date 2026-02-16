package translator

import "strings"

// CSKeyTerms 是中文计算机科学术语到英文的映射字典
var CSKeyTerms = map[string]string{
	// 通用人工智能
	"人工智能": "Artificial Intelligence",
	"AI":      "Artificial Intelligence",
	"机器学习": "Machine Learning",
	"深度学习": "Deep Learning",
	"强化学习": "Reinforcement Learning",
	"联邦学习": "Federated Learning",
	"迁移学习": "Transfer Learning",
	
	// 计算机视觉
	"计算机视觉": "Computer Vision",
	"图像识别":   "Image Recognition",
	"目标检测":   "Object Detection",
	"语义分割":   "Semantic Segmentation",
	
	// 自然语言处理
	"自然语言处理": "Natural Language Processing",
	"大语言模型":   "Large Language Model",
	"LLM":        "Large Language Model",
	"机器翻译":     "Machine Translation",
	"情感分析":     "Sentiment Analysis",
	
	// 系统 / 网络
	"云计算":   "Cloud Computing",
	"边缘计算": "Edge Computing",
	"分布式系统": "Distributed Systems",
	"物联网":   "Internet of Things",
	"IoT":    "Internet of Things",
	"区块链":   "Blockchain",
	"网络安全": "Cybersecurity",
	"5G":     "5G Network",
	"6G":     "6G Network",
	
	// 软件工程
	"软件工程": "Software Engineering",
	"DevOps": "DevOps",
	"微服务":   "Microservices",
	
	// 数据
	"大数据":   "Big Data",
	"数据挖掘": "Data Mining",
	"数据库":   "Database",
	"知识图谱": "Knowledge Graph",
	
	// 理论
	"算法":   "Algorithm",
	"数据结构": "Data Structure",
	"量子计算": "Quantum Computing",
}

// TranslateQuery 尝试将中文查询转换为英文
// 如果未找到翻译则返回原始查询，否则返回翻译后的查询
// 同时返回一个布尔值表示是否发生了翻译
func TranslateQuery(query string) (string, bool) {
	// 首先进行简单查找
	if val, ok := CSKeyTerms[query]; ok {
		return val, true
	}

	// 混合查询的子字符串替换（简单方法）
	// 遍历映射并替换出现的内容
	// 注意：这是一个基本实现。对于生产环境，请使用字典树或更复杂的匹配器。
	translated := query
	hasChange := false
	
	for cn, en := range CSKeyTerms {
		if strings.Contains(translated, cn) {
			translated = strings.ReplaceAll(translated, cn, en)
			hasChange = true
		}
	}

	return translated, hasChange
}

// ContainsChinese 检查字符串是否包含任何中文字符
func ContainsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}
