package utils

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
)

// GenerateFileKey creates a unique object key based on the original
// file name while preserving the extension. Any directory prefix in
// the provided key is kept intact.
func GenerateFileKey(original string) string {
	dir := path.Dir(original)
	base := path.Base(original)
	ext := path.Ext(base)
	name := strings.TrimSuffix(base, ext)
	slug := NormalizeText(name)
	newName := fmt.Sprintf("%s-%s%s", slug, uuid.New().String(), ext)
	if dir == "." {
		return newName
	}
	return path.Join(dir, newName)
}
