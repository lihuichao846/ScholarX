package provider

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

var TopTierVenues = map[string]bool{
	"cvpr": true, "iccv": true, "eccv": true, "neurips": true, "nips": true,
	"icml": true, "iclr": true, "aaai": true, "ijcai": true, "acl": true,
	"emnlp": true, "naacl": true, "siggraph": true, "kdd": true, "www": true,
	"sigmod": true, "vldb": true, "tpami": true, "ijcv": true, "pami": true,
	"tip": true, "tvcg": true, "tog": true, "nature": true, "science": true,
	"cell": true,
}

var CCFCatalog = map[string]string{
	// A Class
	"cvpr": "A", "iccv": "A", "icml": "A", "neurips": "A", "nips": "A", "aaai": "A", "ijcai": "A", "acl": "A",
	"siggraph": "A", "kdd": "A", "www": "A", "sigmod": "A", "vldb": "A", "icde": "A",
	"tpami": "A", "ijcv": "A", "tip": "A", "tvcg": "A", "tog": "A", "tochi": "A",
	"pami": "A", "ieee transactions on pattern analysis and machine intelligence": "A",
	"international journal of computer vision": "A",
	"ieee transactions on image processing": "A",
	"ieee transactions on visualization and computer graphics": "A",
	"acm transactions on graphics": "A",
	"nature": "A", "science": "A", "cell": "A",
	"pldi": "A", "popl": "A", "sosp": "A", "osdi": "A", "asplos": "A", "isca": "A", "micro": "A",
	"sp": "A", "ccs": "A", "uss": "A", "ndss": "A", "crypto": "A", "eurocrypt": "A",

	// B Class
	"eccv": "B", "emnlp": "B", "coling": "B", "naacl": "B", "bmvc": "B", "icme": "B", "icip": "B", "icassp": "B",
	"tmm": "B", "tcsvt": "B", "cviu": "B", "pr": "B",
	"ieee transactions on multimedia": "B",
	"ieee transactions on circuits and systems for video technology": "B",
	"pattern recognition": "B",
	"cikm": "B", "wsdm": "B", "sdm": "B", "icdm": "B",
	"ecai": "B", "uai": "B", "colt": "B",
	"infocom": "B",

	// C Class
	"accv": "C", "icpr": "C", "fg": "C", "wacv": "C",
	"neurocomputing": "C", "iet image processing": "C",
}

func ParseDate(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	// Try ISO
	t, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return t
	}
	// Try YYYY-MM-DD
	t, err = time.Parse("2006-01-02", value)
	if err == nil {
		return t
	}
	return time.Time{}
}

func GetMonthDateRange(yearMonth string) (string, string) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return "", ""
	}
	start := t.Format("2006-01-02")
	// Last day of month
	nextMonth := t.AddDate(0, 1, 0)
	end := nextMonth.Add(-24 * time.Hour).Format("2006-01-02")
	return start, end
}

func GetCCFClass(venue string) string {
	if venue == "" {
		return "None"
	}
	venueLower := strings.ToLower(strings.TrimSpace(venue))
	if val, ok := CCFCatalog[venueLower]; ok {
		return val
	}
	// Sort keys by length descending
	keys := make([]string, 0, len(CCFCatalog))
	for k := range CCFCatalog {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, k := range keys {
		if strings.Contains(venueLower, k) {
			if len(k) < 4 {
				// Regex word boundary check
				matched, _ := regexp.MatchString(`\b`+regexp.QuoteMeta(k)+`\b`, venueLower)
				if matched {
					return CCFCatalog[k]
				}
			} else {
				return CCFCatalog[k]
			}
		}
	}
	return "None"
}

func parseOpenAlexAbstract(inverted map[string][]int) string {
	if len(inverted) == 0 {
		return ""
	}
	maxIndex := 0
	for _, positions := range inverted {
		for _, pos := range positions {
			if pos > maxIndex {
				maxIndex = pos
			}
		}
	}
	words := make([]string, maxIndex+1)
	for word, positions := range inverted {
		for _, pos := range positions {
			if pos < len(words) {
				words[pos] = word
			}
		}
	}
	return strings.Join(words, " ")
}
