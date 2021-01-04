package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	documentStore := LoadDocumentStore()

	index := CreateIndex(documentStore.All())

	//	searcher := Searcher{}
	//	err := searcher.Load("completeworks.txt")
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	//	http.HandleFunc("/search", handleSearch(searcher))
	http.HandleFunc("/search", handleIndexSearch(documentStore, index))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

//type Searcher struct {
//	CompleteWorks string
//	SuffixArray   *suffixarray.Index
//}

type SearchDocument struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func handleIndexSearch(documentStore DocumentStore, index Index) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		// execute query
		queryDocuments := index.QueryConcurrent(query[0])
		// to response
		response := toResponse(queryDocuments, documentStore)

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(&response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func toResponse(queryDocuments []QueryDocument, store DocumentStore) []SearchDocument {
	searchDocuments := make([]SearchDocument, 0, len(queryDocuments))
	for _, qd := range queryDocuments {
		document, _ := store.DocumentById(qd.documentId)
		searchDocuments = append(searchDocuments,
			SearchDocument{
				Id:    qd.documentId,
				Title: document.title,
				Text:  wordRefsText(qd.refs, &document)})
	}
	return searchDocuments
}

func wordRefsText(wordRefs []WordRef, document *Document) string {
	tokensHandled := make(map[string]int, 10)
	var sb strings.Builder
	seperator := "..."
	for i, wordRef := range wordRefs {
		count, ok := tokensHandled[wordRef.token]
		if !ok {
			count = 0
		}
		if count > 3 {
			continue
		}
		tokensHandled[wordRef.token] = count + 1
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(seperator)
		text := wordRefText(wordRef, document)
		sb.WriteString(text)
	}
	sb.WriteString(" ")
	sb.WriteString(seperator)
	return sb.String()
}

func wordRefText(wordRef WordRef, document *Document) string {
	p := 40
	startPos := Max(0, wordRef.end-p)
	endPos := Min(len(document.text), wordRef.end+40)
	return fmt.Sprintf("%s<span class=\"hl\">%s</span>%s",
		(*document).text[startPos:wordRef.start],
		(*document).text[wordRef.start:wordRef.end],
		(*document).text[wordRef.end:endPos])
}

func Max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

//func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		query, ok := r.URL.Query()["q"]
//		if !ok || len(query[0]) < 1 {
//			w.WriteHeader(http.StatusBadRequest)
//			w.Write([]byte("missing search query in URL params"))
//			return
//		}
//		results := searcher.Search(query[0])
//		buf := &bytes.Buffer{}
//		enc := json.NewEncoder(buf)
//		err := enc.Encode(results)
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			w.Write([]byte("encoding failure"))
//			return
//		}
//		w.Header().Set("Content-Type", "application/json")
//		w.Write(buf.Bytes())
//	}
//}

//func (s *Searcher) Load(filename string) error {
//	dat, err := ioutil.ReadFile(filename)
//	if err != nil {
//		return fmt.Errorf("Load: %w", err)
//	}
//	s.CompleteWorks = string(dat)
//	s.SuffixArray = suffixarray.New(dat)
//	return nil
//}

//func (s *Searcher) Search(query string) []string {
//	idxs := s.SuffixArray.Lookup([]byte(query), -1)
//	results := []string{}
//	for _, idx := range idxs {
//		results = append(results, s.CompleteWorks[idx-250:idx+250])
//	}
//	return results
//}
