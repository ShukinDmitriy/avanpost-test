package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"word-search-in-files/pkg/searcher"
)

type App struct {
	searcher *searcher.Searcher
}

func NewApp(searcher *searcher.Searcher) App {
	return App{
		searcher: searcher,
	}
}

func (app *App) SearchHandler(w http.ResponseWriter, r *http.Request) {
	res := &SearchResponse{
		List:  nil,
		Error: "",
	}

	qParams := r.URL.Query()["q"]

	if len(qParams) == 0 {
		res.Error = "No search string"
		sendJSON(w, res, http.StatusNotFound)
		return
	}

	searchWord := qParams[0]
	if searchWord == "" {
		res.Error = "Empty search string"
		sendJSON(w, res, http.StatusNotFound)
		return
	}

	files, err := app.searcher.Search(searchWord)
	if err != nil {
		res.Error = err.Error()
		sendJSON(w, res, http.StatusNotFound)
		return
	}

	responseCode := http.StatusNotFound
	if len(files) > 0 {
		res.List = files
		responseCode = http.StatusOK
	}

	sendJSON(w, res, responseCode)
}

type SearchResponse struct {
	List  []string `json:"list"`
	Error string   `json:"error"`
}

func sendJSON(w http.ResponseWriter, response *SearchResponse, responseCode int) {
	jsonRes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	fmt.Fprint(w, string(jsonRes))
}
