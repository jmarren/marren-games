package db

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type WithQueriesMap struct {
	queries *map[string]string
}

func CreateWithQueriesMap() *WithQueriesMap {
	// get current working directory
	cwd, _ := os.Getwd()
	fmt.Println("current directory: ", cwd)

	sqlDir := os.DirFS(cwd + "/internal/sql")
	// fmt.Println("sqlDir: ", sqlDir)

	sqlFiles, err := fs.ReadDir(sqlDir, ".")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sqlFiles)

	withQueries := make(map[string]string)

	for _, file := range sqlFiles {
		query := ""
		reader, err := fs.ReadFile(sqlDir, file.Name())
		if err != nil {
			fmt.Println(err)
		}
		bytesReader := bytes.NewReader(reader)
		scanner := bufio.NewScanner(bytesReader)

		for scanner.Scan() {
			query += scanner.Text()
		}
		queryName := strings.Trim(file.Name(), ".sql")
		withQueries[queryName] = query
	}

	fmt.Println("----- WithQueries parsed from sql directory: ")
	fmt.Println(withQueries)

	return &WithQueriesMap{queries: &withQueries}
}

func (w *WithQueriesMap) GetWithQuery(queryName string) string {
	return (*w.queries)[queryName]
}
