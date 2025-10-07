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

	// Загружаем примеры данных
	pipeline.vectorStore.LoadSampleData()

	return pipeline
}

func (p *RAGPipeline) ProcessQuery(question string) (string, []Document) {
	// Ищем релевантные документы
	similarDocs := p.vectorStore.SearchSimilar(question, 3)

	if len(similarDocs) == 0 {
		return "Релевантная информация не найдена в базе знаний.", similarDocs
	}

	// Формируем контекст для промпта
	context := p.buildContext(similarDocs)

	return context, similarDocs
}

func (p *RAGPipeline) buildContext(docs []Document) string {
	if len(docs) == 0 {
		return ""
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Релевантная информация из базы знаний:\n\n")

	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc.Content))
	}

	contextBuilder.WriteString("\nИспользуй эту информацию для ответа на вопрос.")
	return contextBuilder.String()
}

func (p *RAGPipeline) AddDocument(content string) string {
	return p.vectorStore.AddDocument(content)
}

func (p *RAGPipeline) GetStats() map[string]interface{} {
	return p.vectorStore.GetStats()
}
