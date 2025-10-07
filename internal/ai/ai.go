package ai

type AIClient interface {
	Ask(question string) (string, error)
}
