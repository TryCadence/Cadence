package patterns

type WebPatternStrategy interface {
	Name() string
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

func (r *WebPatternRegistry) GetStrategies() []WebPatternStrategy {
	return r.strategies
}
