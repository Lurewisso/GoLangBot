package rag

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

type Document struct {
	ID      string
	Content string
	Vector  []float64
}

type VectorStore struct {
	documents []Document
	mu        sync.RWMutex
}

func NewVectorStore() *VectorStore {
	return &VectorStore{
		documents: make([]Document, 0),
	}
}

// Простая реализация эмбеддингов через TF-IDF like подход
func (vs *VectorStore) textToVector(text string) []float64 {
	words := strings.Fields(strings.ToLower(text))
	wordFreq := make(map[string]float64)

	for _, word := range words {
		// Убираем пунктуацию и короткие слова
		word = strings.Trim(word, ".,!?;:\"'")
		if len(word) > 2 {
			wordFreq[word]++
		}
	}

	// Нормализуем частоты
	maxFreq := 0.0
	for _, freq := range wordFreq {
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	vector := make([]float64, len(wordFreq))
	i := 0
	for _, freq := range wordFreq {
		vector[i] = freq / maxFreq
		i++
	}

	return vector
}

// Косинусное сходство
func cosineSimilarity(a, b []float64) float64 {
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a) && i < len(b); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (vs *VectorStore) AddDocument(content string) string {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	id := fmt.Sprintf("doc_%d", len(vs.documents))
	vector := vs.textToVector(content)

	doc := Document{
		ID:      id,
		Content: content,
		Vector:  vector,
	}

	vs.documents = append(vs.documents, doc)
	return id
}

func (vs *VectorStore) SearchSimilar(query string, topK int) []Document {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if len(vs.documents) == 0 {
		return []Document{}
	}

	queryVector := vs.textToVector(query)

	type scoredDoc struct {
		doc   Document
		score float64
	}

	scoredDocs := make([]scoredDoc, 0, len(vs.documents))

	for _, doc := range vs.documents {
		score := cosineSimilarity(queryVector, doc.Vector)
		if score > 0.1 { // Минимальный порог сходства
			scoredDocs = append(scoredDocs, scoredDoc{doc: doc, score: score})
		}
	}

	// Сортируем по убыванию сходства
	for i := 0; i < len(scoredDocs); i++ {
		for j := i + 1; j < len(scoredDocs); j++ {
			if scoredDocs[j].score > scoredDocs[i].score {
				scoredDocs[i], scoredDocs[j] = scoredDocs[j], scoredDocs[i]
			}
		}
	}

	// Берем topK результатов
	if len(scoredDocs) > topK {
		scoredDocs = scoredDocs[:topK]
	}

	result := make([]Document, len(scoredDocs))
	for i, sd := range scoredDocs {
		result[i] = sd.doc
	}

	return result
}

func (vs *VectorStore) LoadSampleData() {
	sampleData := []string{
		"RAG (Retrieval-Augmented Generation) - это архитектура, которая сочетает поиск информации и генерацию текста.",
		"RAG сначала ищет релевантные документы в базе знаний, затем использует их для генерации ответа.",
		"Векторный поиск позволяет находить семантически похожие тексты даже без точного совпадения слов.",
		"Telegram боты создаются через BotFather и используют API для отправки сообщений.",
		"Go (Golang) - статически типизированный язык программирования с сборщиком мусора и поддержкой многопоточности.",
		"Docker позволяет упаковывать приложения в контейнеры для удобного развертывания.",
		"API ключи необходимы для доступа к сервисам искусственного интеллекта как DeepSeek и OpenRouter.",
	}

	for _, content := range sampleData {
		vs.AddDocument(content)
	}
}

func (vs *VectorStore) GetStats() map[string]interface{} {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	return map[string]interface{}{
		"total_documents": len(vs.documents),
		"store_size":      fmt.Sprintf("%d docs", len(vs.documents)),
	}
}
