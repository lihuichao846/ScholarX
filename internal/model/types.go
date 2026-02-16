package model

// --- 数据模型 ---

type Paper struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Venue       string   `json:"venue"`
	Year        *int     `json:"year"`
	Abstract    string   `json:"abstract"`
	URL         string   `json:"url"`
	Source      string   `json:"source"`
	Categories  []string `json:"categories"`
	PublishedAt string   `json:"published_at"`
	Citations   int      `json:"citations"`
	CCFClass    string   `json:"ccf_class"`
}

type PaperResponse struct {
	Count       int     `json:"count"`
	Items       []Paper `json:"items"`
	Translation string  `json:"translation,omitempty"`
}

// --- ArXiv XML Structs ---

type AtomFeed struct {
	Entries []AtomEntry `xml:"entry"`
}

type AtomEntry struct {
	ID        string       `xml:"id"`
	Published string       `xml:"published"`
	Title     string       `xml:"title"`
	Summary   string       `xml:"summary"`
	Authors   []AtomAuthor `xml:"author"`
	Links     []AtomLink   `xml:"link"`
	Category  []AtomCat    `xml:"category"`
	// ArXiv specific fields need namespace handling
	Comment    string `xml:"http://arxiv.org/schemas/atom comment"`
	JournalRef string `xml:"http://arxiv.org/schemas/atom journal_ref"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type AtomCat struct {
	Term string `xml:"term,attr"`
}

// --- OpenAlex JSON 结构体 ---

type OAWork struct {
	ID               string           `json:"id"`
	DisplayName      string           `json:"display_name"`
	Authorships      []OAAuthorship   `json:"authorships"`
	PublicationYear  int              `json:"publication_year"`
	PublicationDate  string           `json:"publication_date"`
	Concepts         []OAConcept      `json:"concepts"`
	PrimaryLocation  OALocation       `json:"primary_location"`
	CitedByCount     int              `json:"cited_by_count"`
	AbstractInverted map[string][]int `json:"abstract_inverted_index"`
}

type OAAuthorship struct {
	Author struct {
		DisplayName string `json:"display_name"`
	} `json:"author"`
}

type OAConcept struct {
	DisplayName string `json:"display_name"`
}

type OALocation struct {
	Source struct {
		DisplayName string `json:"display_name"`
	} `json:"source"`
	LandingPageURL string `json:"landing_page_url"`
}

type OAResponse struct {
	Results []OAWork `json:"results"`
}
