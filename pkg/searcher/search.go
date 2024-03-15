package searcher

import (
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"word-search-in-files/pkg/internal/dir"
)

type Index struct {
	mx sync.RWMutex
	m  map[string][]string
}

func (c *Index) Load(key string) ([]string, bool) {
	c.mx.Lock()
	val, ok := c.m[key]
	c.mx.Unlock()
	return val, ok
}

func (c *Index) Store(key string, value []string) {
	c.mx.Lock()
	c.m[key] = value
	c.mx.Unlock()
}

type Searcher struct {
	index *Index
	FS    fs.FS
}

func (s *Searcher) parseFile(filename string) {
	nameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
	file, err := s.FS.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var size int
	if info, err := file.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}

	fileContent := make([]byte, 0, size+1)
	for {
		if len(fileContent) >= cap(fileContent) {
			d := append(fileContent[:cap(fileContent)], 0)
			fileContent = d[:len(fileContent)]
		}
		n, err := file.Read(fileContent[len(fileContent):cap(fileContent)])
		fileContent = fileContent[:len(fileContent)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	// читаем файл и разбиваем его на слова
	// тут нужно еще применять нормализацию, токенизацию и прочий морфологический анализ,
	// который превращает тестовое задание в написание индексатора и поискового движка...
	// Также не очень понятно, что нужно вырезать из текста, т.к. по тесту World1 это прям такое слово
	regex := regexp.MustCompile("[.,!\"'?()]")
	content := regex.ReplaceAllString(string(fileContent), " ")
	regex = regexp.MustCompile("\\s{2,}")
	content = strings.ToLower(regex.ReplaceAllString(content, " "))
	words := strings.Split(content, " ")

	for _, word := range words {
		cur, ok := s.index.Load(word)
		if !ok {
			cur = make([]string, 0)
		}
		if slices.Index(cur, nameWithoutExtension) < 0 {
			cur = append(cur, nameWithoutExtension)
			slices.Sort(cur)
			s.index.Store(word, cur)
		}
	}
}

// Init чтобы обеспечить O(1) для поиска необходимо создать индекс, а не выполнять поиск каждый раз
func (s *Searcher) Init() error {
	s.index = &Index{
		m: make(map[string][]string),
	}

	filenames, err := dir.FilesFS(s.FS, "")
	if err != nil {
		return err
	}

	// в требованиях было про параллельный поиск, но в индекс сохраняется список файлов для конкретного слова
	// поэтому можно просто прочитать и обработать файлы в параллель
	wg := sync.WaitGroup{}
	for _, filename := range filenames {
		wg.Add(1)
		go func(f string) {
			s.parseFile(f)
			wg.Done()
		}(filename)
	}

	wg.Wait()

	return nil
}

func (s *Searcher) Search(word string) (files []string, err error) {
	res, _ := s.index.Load(strings.ToLower(word))

	return res, nil
}
