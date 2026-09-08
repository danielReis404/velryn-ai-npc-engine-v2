package engine

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
