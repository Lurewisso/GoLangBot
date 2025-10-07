package rag

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

type Document struct {
	ID      string
	Content string
	Vector  []float64
	Tokens  []string
}

type VectorStore struct {
	documents  []Document
	vocabulary map[string]int
	docVectors [][]float64
	mu         sync.RWMutex
}

func NewVectorStore() *VectorStore {
	return &VectorStore{
		documents:  make([]Document, 0),
		vocabulary: make(map[string]int),
		docVectors: make([][]float64, 0),
	}
}

func (vs *VectorStore) tokenize(text string) []string {
	text = strings.ToLower(text)

	reg := regexp.MustCompile(`[^\w\sа-яё]`)

	text = reg.ReplaceAllString(text, " ")

	words := strings.Fields(text)

	var tokens []string
	stopWords := map[string]bool{
		"и": true, "в": true, "на": true, "с": true, "по": true, "для": true,
		"не": true, "что": true, "это": true, "как": true, "так": true,
		"из": true, "у": true, "к": true, "о": true, "за": true, "от": true,
		"то": true, "же": true, "все": true, "но": true, "вы": true, "бы": true,
		"а": true, "мне": true, "вот": true, "до": true, "ну": true, "ли": true,
		"если": true, "уже": true, "или": true, "ни": true, "быть": true, "был": true,
		"про": true, "при": true, "год": true, "очень": true, "может": true, "есть": true,
	}

	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) >= 2 && !stopWords[word] && isRussianWord(word) {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

func isRussianWord(word string) bool {
	russianCount := 0
	totalCount := 0

	for _, r := range word {
		if unicode.Is(unicode.Cyrillic, r) {
			russianCount++
		}
		totalCount++
	}

	return russianCount > 0 && float64(russianCount)/float64(totalCount) > 0.6
}

func (vs *VectorStore) buildVocabulary() {
	vs.vocabulary = make(map[string]int)
	index := 0

	for _, doc := range vs.documents {
		for _, token := range doc.Tokens {
			if _, exists := vs.vocabulary[token]; !exists {
				vs.vocabulary[token] = index
				index++
			}
		}
	}
}

func (vs *VectorStore) textToVector(tokens []string) []float64 {
	if len(vs.vocabulary) == 0 {
		return []float64{}
	}

	vector := make([]float64, len(vs.vocabulary))

	tf := make(map[string]float64)
	for _, token := range tokens {
		if idx, exists := vs.vocabulary[token]; exists {
			tf[token]++
			vector[idx] = tf[token]
		}
	}

	maxFreq := 0.0
	for _, freq := range tf {
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	if maxFreq > 0 {
		for token, freq := range tf {
			if idx, exists := vs.vocabulary[token]; exists {
				vector[idx] = freq / maxFreq
			}
		}
	}

	return vector
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
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
	tokens := vs.tokenize(content)

	doc := Document{
		ID:      id,
		Content: content,
		Tokens:  tokens,
	}

	vs.documents = append(vs.documents, doc)

	vs.buildVocabulary()
	vs.updateAllVectors()

	return id
}

func (vs *VectorStore) updateAllVectors() {
	vs.docVectors = make([][]float64, len(vs.documents))
	for i, doc := range vs.documents {
		vs.docVectors[i] = vs.textToVector(doc.Tokens)
	}
}

func (vs *VectorStore) SearchSimilar(query string, topK int) []Document {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if len(vs.documents) == 0 {
		return []Document{}
	}

	queryTokens := vs.tokenize(query)
	queryVector := vs.textToVector(queryTokens)

	type scoredDoc struct {
		doc   Document
		score float64
	}

	scoredDocs := make([]scoredDoc, 0, len(vs.documents))

	for i, docVector := range vs.docVectors {
		score := cosineSimilarity(queryVector, docVector)

		if score > 0.05 {
			scoredDocs = append(scoredDocs, scoredDoc{
				doc:   vs.documents[i],
				score: score,
			})
		}
	}

	for i := 0; i < len(scoredDocs); i++ {
		for j := i + 1; j < len(scoredDocs); j++ {
			if scoredDocs[j].score > scoredDocs[i].score {
				scoredDocs[i], scoredDocs[j] = scoredDocs[j], scoredDocs[i]
			}
		}
	}

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
		"RAG Retrieval Augmented Generation архитектура поиск информация генерация текст",
		"RAG сначала ищет релевантные документы базе знаний затем использует генерацию ответа",
		"Векторный поиск позволяет находить семантически похожие тексты точное совпадение слов",
		"Telegram боты создаются BotFather используют API отправки сообщений",
		"Go Golang статически типизированный язык программирования сборщик мусора поддержка многопоточности",
		"Docker позволяет упаковывать приложения контейнеры удобное развертывание",
		"API ключи необходимы доступа сервисам искусственного интеллекта DeepSeek OpenRouter",
		"Программирование разработка программ обеспечение компьютеров алгоритмы код",
		"Искусственный интеллект AI машинное обучение нейронные сети данные обучение модели",
		"База данных хранение информации структурированные данные запросы SQL",
		"Веб разработка создание сайтов приложений интерфейсы backend frontend",
		"Мобильные приложения iOS Android разработка телефоны планшеты",
		"Облачные вычисления сервера хранение данных AWS Google Cloud Azure",
		"Блокчейн криптовалюты Bitcoin Ethereum смарт контракты децентрализация",
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
		"vocabulary_size": len(vs.vocabulary),
		"store_size":      fmt.Sprintf("%d docs, %d words", len(vs.documents), len(vs.vocabulary)),
	}
}
