package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	cli "github.com/jawher/mow.cli"
)

// I18nItem hold i18n info for a key
type I18nItem struct {
	key      string
	value    string
	filename string
}

// Migrator migrate i18n keys from one file to another
type Migrator struct {
	src    []string
	dst    string
	keys   []string
	result chan *I18nItem
	wg     *sync.WaitGroup
}

// NewMigrator creates a migrator instance
func NewMigrator(s []string, d string, k io.Reader) *Migrator {
	return &Migrator{
		src:    s,
		dst:    d,
		keys:   readKeys(k),
		result: make(chan *I18nItem),
		wg:     new(sync.WaitGroup),
	}
}

// Run launches the migration
func (m *Migrator) Run() {
	// filename->key->val map
	i18nMultiMap := make(map[string]map[string]string)

	// run the workers to search for query
	for _, src := range m.src {
		filepath.Walk(src, func(path string, file os.FileInfo, err error) error {
			if !file.IsDir() {
				m.wg.Add(1)
				go m.searchInFile(path)
			}
			return nil
		})
	}

	// wait for workers and then close the result chan
	go func() {
		m.wg.Wait()
		close(m.result)
	}()

	// consume the result chan
	for item := range m.result {
		_, ok := i18nMultiMap[item.filename]
		if !ok {
			i18nMultiMap[item.filename] = make(map[string]string)
		}
		i18nMultiMap[item.filename][item.key] = item.value
	}

	// write files
	for filename, i18nMap := range i18nMultiMap {
		if len(i18nMap) == 0 {
			continue
		}
		f, err := os.Create(fmt.Sprintf("%s/%s", m.dst, filename))
		defer f.Close()
		check(err)

		for _, key := range m.keys {
			if len(key) == 0 || strings.HasPrefix(key, "#") {
				f.WriteString(fmt.Sprintln(key))
				continue
			}

			val, ok := i18nMap[key]
			if ok {
				f.WriteString(fmt.Sprintln(val))
			}
		}
	}
}

func (m *Migrator) searchInFile(path string) {
	defer m.wg.Done()
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		for _, key := range m.keys {
			if len(key) == 0 {
				continue
			}
			if strings.HasPrefix(scanner.Text(), key) {
				item := I18nItem{
					key:      key,
					value:    scanner.Text(),
					filename: filepath.Base(path),
				}
				m.result <- &item
			}
		}
	}
}

func readKeys(keysFile io.Reader) []string {
	result := []string{}
	scanner := bufio.NewScanner(keysFile)
	for i := 1; scanner.Scan(); i++ {
		result = append(result, scanner.Text())
	}
	return result
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkFilesExists(paths []string) {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("%s does not exist\n", path)
			os.Exit(1)
		}
	}
}

func main() {
	app := cli.App("i18n_migrate", "Migrate I18n files")
	app.ErrorHandling = flag.ExitOnError
	app.Spec = "SRC... DST KEYS"
	src := app.StringsArg("SRC", nil, "Source files to copy")
	dst := app.StringArg("DST", "", "Destination where to copy files to")
	keys := app.StringArg("KEYS", "", "File containing i18n keys you want to copy")

	app.Action = func() {
		paths := append(*src, *dst)
		paths = append(paths, *keys)
		checkFilesExists(paths)

		keysFile, err := os.Open(*keys)
		defer keysFile.Close()
		if err != nil {
			panic(fmt.Sprintf("file %s not found", keys))
		}

		m := NewMigrator(*src, *dst, keysFile)
		m.Run()
	}

	app.Run(os.Args)
}
