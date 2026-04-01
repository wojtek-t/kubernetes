package main

import (
	"os"
	"strings"
)

func main() {
	path := "pkg/apis/scheduling/types.go"
	contentBytes, err := os.ReadFile(path)
	if err == nil {
		content := string(contentBytes)
		content = strings.ReplaceAll(content, "	Priority *int32\n\n}", "	Priority *int32\n}")
		// Wait, the grep shows it's already `Priority *int32\n}`!
		// Let's check `git diff origin/master...HEAD pkg/apis/scheduling/types.go`
	}
}
