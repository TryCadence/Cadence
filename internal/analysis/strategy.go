package analysis

// Strategy categories for classification of detection strategies.
const (
	CategoryVelocity      = "velocity"
	CategoryStructural    = "structural"
	CategoryBehavioral    = "behavioral"
	CategoryStatistical   = "statistical"
	CategoryPattern       = "pattern"
	CategoryLinguistic    = "linguistic"
	CategoryAccessibility = "accessibility"
)

// StrategyInfo provides serializable metadata about a detection strategy.
// This is the common representation used in reports, registries, and APIs.
type StrategyInfo struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Confidence  float64  `json:"confidence"`
	Description string   `json:"description"`
	SourceTypes []string `json:"source_types"`
}
