package patterns

// WebPatternStrategy is the interface for all web-based detection strategies.
// Each strategy must provide metadata (Category, Confidence, Description)
// in addition to its detection logic.
type WebPatternStrategy interface {
	Name() string
	Category() string
	Confidence() float64
	Description() string
	Detect(content string, wordCount int) *DetectionResult
}

type DetectionResult struct {
	Detected    bool
	Type        string
	Severity    float64
	Description string
	Examples    []string
}

type WebPatternRegistry struct {
	strategies []WebPatternStrategy
}

func NewWebPatternRegistry() *WebPatternRegistry {
	registry := &WebPatternRegistry{
		strategies: make([]WebPatternStrategy, 0),
	}

	registry.RegisterDefaults()
	return registry
}

func (r *WebPatternRegistry) Register(strategy WebPatternStrategy) {
	r.strategies = append(r.strategies, strategy)
}

func (r *WebPatternRegistry) RegisterDefaults() {
	r.Register(NewOverusedPhrasesStrategy())
	r.Register(NewGenericLanguageStrategy())
	r.Register(NewExcessiveStructureStrategy())
	r.Register(NewPerfectGrammarStrategy())
	r.Register(NewBoilerplateTextStrategy())
	r.Register(NewRepetitivePatternsStrategy())
	r.Register(NewMissingNuanceStrategy())
	r.Register(NewExcessiveTransitionsStrategy())
	r.Register(NewUniformSentenceLengthStrategy())
	r.Register(NewAIVocabularyStrategy())
	r.Register(NewEmojiStrategy())
	r.Register(NewSpecialCharactersStrategy())
	r.Register(NewMissingAltTextStrategy())
	r.Register(NewSemanticHTMLStrategy())
	r.Register(NewAccessibilityMarkersStrategy())
	r.Register(NewHeadingHierarchyStrategy())
	r.Register(NewHardcodedValuesStrategy())
	r.Register(NewFormIssuesStrategy())
	r.Register(NewLinkTextQualityStrategy())
	r.Register(NewGenericStylingStrategy())
}

func (r *WebPatternRegistry) DetectAll(content string, wordCount int) []*DetectionResult {
	results := make([]*DetectionResult, 0)

	for _, strategy := range r.strategies {
		if result := strategy.Detect(content, wordCount); result != nil && result.Detected {
			results = append(results, result)
		}
	}

	return results
}

func (r *WebPatternRegistry) DetectAllWithPassed(content string, wordCount int) []*DetectionResult {
	results := make([]*DetectionResult, 0)

	for _, strategy := range r.strategies {
		result := strategy.Detect(content, wordCount)
		if result == nil {
			result = &DetectionResult{
				Detected:    false,
				Type:        strategy.Name(),
				Severity:    0,
				Description: "No issues detected",
				Examples:    nil,
			}
		} else if !result.Detected {
			if result.Type == "" {
				result.Type = strategy.Name()
			}
			if result.Description == "" {
				result.Description = "No issues detected"
			}
		}
		results = append(results, result)
	}

	return results
}

func (r *WebPatternRegistry) GetStrategies() []WebPatternStrategy {
	return r.strategies
}
