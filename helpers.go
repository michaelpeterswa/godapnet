package godapnet

func sliceStringByN(text string, max int) []string {
	var texts []string
	textLength := len(text)
	splits := (textLength / max) + 1

	for i := 1; i < splits+1; i++ {
		j := i * max
		lower := j - max
		upper := j

		if i < splits {
			texts = append(texts, text[lower:upper])
		} else if text[lower:textLength] != "" {
			texts = append(texts, text[lower:textLength])
		}
	}
	return texts
}
