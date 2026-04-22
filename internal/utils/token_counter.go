package utils

// CountTokens provides a rough token estimation (1 token ~= 4 characters).
// Production should use a library like tiktoken for accurate counts.
func CountTokens(text string) int64 {
	if text == "" {
		return 0
	}
	return int64(len(text) / 4)
}

// CountMessageTokens counts tokens across a slice of message maps.
func CountMessageTokens(messages []map[string]string) int64 {
	var total int64
	for _, msg := range messages {
		for _, value := range msg {
			total += CountTokens(value)
		}
	}
	// Add overhead for message formatting
	total += int64(len(messages) * 4)
	return total
}

// CountResponseTokens counts tokens in a response string.
func CountResponseTokens(content string) int64 {
	return CountTokens(content)
}
