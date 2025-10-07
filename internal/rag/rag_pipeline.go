package rag

import (
	"fmt"
	"strings"
)

type RAGPipeline struct {
	vectorStore *VectorStore
}

func NewRAGPipeline() *RAGPipeline {
	pipeline := &RAGPipeline{
		vectorStore: NewVectorStore(),
	}

	pipeline.vectorStore.LoadSampleData()

	return pipeline
}

func (p *RAGPipeline) ProcessQuery(question string) (string, []Document) {

	similarDocs := p.vectorStore.SearchSimilar(question, 5)

	if len(similarDocs) == 0 {
		return "", similarDocs
	}

	context := p.buildContext(similarDocs)

	return context, similarDocs
}

func (p *RAGPipeline) buildContext(docs []Document) string {
	if len(docs) == 0 {
		return ""
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Контекст для ответа:\n\n")

	for _, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("- %s\n", doc.Content))
	}

	contextBuilder.WriteString("\nИспользуй эту информацию для формирования ответа.")
	return contextBuilder.String()
}

func (p *RAGPipeline) AddDocument(content string) string {
	return p.vectorStore.AddDocument(content)
}

func (p *RAGPipeline) GetStats() map[string]interface{} {
	return p.vectorStore.GetStats()
}
