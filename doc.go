package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	root := "./commands" // Start from the commands directory
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			fmt.Printf("File: %s\n", path)
			for _, decl := range node.Decls {
				// Process each declaration (functions, types, etc.)
				// Extract and document key functionalities
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %v\n", err)
	}
}
