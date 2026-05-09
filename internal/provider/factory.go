package provider

type OllamaFactory struct {
	name     string
	endpoint string
}

func NewOllamaFactory(name string, endpoint string) *OllamaFactory {
	return &OllamaFactory{
		name:     name,
		endpoint: endpoint,
	}
}

func (f *OllamaFactory) Name() string { return f.name }

func (f *OllamaFactory) Create(model string, temperature float64) Provider {
	return NewOllamaWithTemp(f.name, f.endpoint, model, temperature)
}

type OpenCodeFactory struct {
	name     string
	endpoint string
	apiKey   string
}

func NewOpenCodeFactory(name string, endpoint, apiKey string) *OpenCodeFactory {
	return &OpenCodeFactory{
		name:     name,
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

func (f *OpenCodeFactory) Name() string { return f.name }

func (f *OpenCodeFactory) Create(model string, temperature float64) Provider {
	return NewOpenCodeWithTemp(f.name, f.endpoint, model, f.apiKey, temperature)
}

type LMStudioFactory struct {
	name     string
	endpoint string
}

func NewLMStudioFactory(name string, endpoint string) *LMStudioFactory {
	return &LMStudioFactory{
		name:     name,
		endpoint: endpoint,
	}
}

func (f *LMStudioFactory) Name() string { return f.name }

func (f *LMStudioFactory) Create(model string, temperature float64) Provider {
	return NewLMStudioWithTemp(f.name, f.endpoint, model, temperature)
}
