package linter

type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
)

type Category string

const (
	Indexability     Category = "indexability"
	Links            Category = "links"
	OnPage           Category = "on_page"
	Redirects        Category = "redirects"
	Internal         Category = "internal"
	SearchTraffic    Category = "search_traffic"
	XMLSitemaps      Category = "xml_sitemaps"
	Security         Category = "security"
	International    Category = "international"
	Accessibility    Category = "accessibility"
	AMP              Category = "amp"
	DuplicateContent Category = "duplicate_content"
	MobileFriendly   Category = "mobile_friendly"
	Performance      Category = "performance"
	Rendered         Category = "rendered"
)

type Tag string

const (
	Issue          Tag = "issue"
	Opportunity    Tag = "opportunity"
	PotentialIssue Tag = "potential_issue"
)
