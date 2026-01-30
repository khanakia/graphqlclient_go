package util

import (
	"fmt"
	"os"

	"github.com/ubgo/goutil"
	"github.com/vektah/gqlparser/v2/ast"
)

func Errorf(pos *ast.Position, msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}

func DumpStructToFile(v interface{}, filename string) error {
	content, err := goutil.ToJSONIndent(v)
	if err != nil {
		return fmt.Errorf("failed to dump struct to file: %w", err)
	}
	return SaveToFile(filename, content)
}

func SaveToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
