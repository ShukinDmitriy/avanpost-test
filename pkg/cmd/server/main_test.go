package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"word-search-in-files/pkg/internal/app"
	"word-search-in-files/pkg/searcher"
)

func TestSearchHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		name  string
		query string
		want
	}{
		{
			name:  "Negative: no query",
			query: "",
			want: want{
				code:        404,
				contentType: "application/json",
				response:    `{"list": null, "error": "No search string"}`,
			},
		},
		{
			name:  "Negative: Empty query",
			query: "q=",
			want: want{
				code:        404,
				contentType: "application/json",
				response:    `{"list": null, "error": "Empty search string"}`,
			},
		},
		{
			name:  "Negative: Empty list",
			query: "q=qweqweqwe",
			want: want{
				code:        404,
				contentType: "application/json",
				response:    `{"list": null, "error": ""}`,
			},
		},
		{
			name:  "Positive: world",
			query: "q=world",
			want: want{
				code:        200,
				contentType: "application/json",
				response:    `{"list": ["file1", "file3"], "error": ""}`,
			},
		},
	}

	fs := fstest.MapFS{
		"file1.txt": {Data: []byte("World")},
		"file2.txt": {Data: []byte("World1")},
		"file3.txt": {Data: []byte("Hello World")},
	}

	s := &searcher.Searcher{
		FS: fs,
	}
	s.Init()
	a := app.NewApp(s)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/files/search"

			if len(test.query) > 0 {
				url += "?" + test.query
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("Error creating new request: %v", err)
			}

			w := httptest.NewRecorder()

			a.SearchHandler(w, req)

			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.JSONEq(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}

}
