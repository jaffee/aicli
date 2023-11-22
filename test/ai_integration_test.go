package test

import (
	"os"
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/jaffee/aicli/pkg/ollama"
	"github.com/jaffee/aicli/pkg/openai"
	"github.com/stretchr/testify/assert"
)

func TestAIs(t *testing.T) {
	AIs := make(map[string]aicli.AI)
	AIs["echo"] = &aicli.Echo{}
	if key := os.Getenv("OPENAI_KEY"); key != "" {
		conf := openai.NewConfig()
		conf.APIKey = key
		AIs["openai"] = openai.NewClient(conf)
	}
	if ollamaHost := os.Getenv("OLLAMA_HOST"); ollamaHost != "" {
		conf := ollama.NewConfig()
		conf.Host = ollamaHost
		AIs["ollama"] = ollama.NewClient(conf)
	}

	for name, ai := range AIs {
		t.Run(name, func(t *testing.T) {
			GetEmbeddingTest(t, ai, models[name])
		})
	}

}

// TODO: this goes away once we can get a list of available embedding models from the AI
var models = map[string]string{
	"openai": "text-embedding-ada-002",
	"echo":   "",
	"ollama": "llama2:nogpu",
}

func GetEmbeddingTest(t *testing.T, ai aicli.AI, model string) {
	req := &aicli.EmbeddingRequest{
		Model:  model,
		Inputs: []string{"Hello, please embed this"},
	}
	embs, err := ai.GetEmbedding(req)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, 1, len(embs)) {
		return
	}
	if !assert.True(t, len(embs[0].Embedding) > 0) {
		return
	}
	if !assert.NotEqual(t, 0.0, embs[0].Embedding[0]) {
		return
	}
	t.Logf("%+v", embs[0].Embedding)
}
