package main

import "fmt"
import "io/ioutil"
import "log"
import "os"
import "strings"
import "unicode"
import snowballeng "github.com/kljensen/snowball/english"
import "sort"
import "sync"

// Index is an inverted map based alogorithm for fast full-text search based on tokens.
//
// The Index analyses documents as follows:
//
//   1. tokenize documents (extract words)
//
//   2. apply filters to each token
//
//      2.1. lower case filter
//
//      2.2. stopword filter (remove common words, see stopwords_en.txt)
//
//      2.3. apply stemmer (normalize forms of the same word, e.g. fish, fishes -> fish)
//
// The index stores the output of the analyzer. The data structure is defined as follows:
//
//   * Inverted Map: map[token] [] DocRef(*document, start, end) ]
//
//   * Reference to a word withn a document: DocRef(*document, start, end)
//
// A query against the index works as follows:
//
//   1. analyze query term to get search tokens
//
// 	 2. lookup index for all analyzed tokens
//
//   3. sort (rank) search results
//
//      * sort by distinct token count desc
//      * if token count is equal for given documents, sort by ref count desc
//

//func x_main() {
//
//	documents := ParseDocuments()
//
//	index := CreateIndex(documents)
//
//	queryDocuments := index.Query("Nile William")
//	for _, queryDocument := range queryDocuments {
//		fmt.Printf("Document found: %s %d\n", queryDocument.documentId, len(queryDocument.refs))
//	}
//	queryDocuments = index.QueryConcurrent("Nile William")
//	for _, queryDocument := range queryDocuments {
//		fmt.Printf("Document found: %s %d\n", queryDocument.documentId, len(queryDocument.refs))
//	}
//}

// Positional Token within text
type Token struct {
	value string
	start int
	end   int
}

// A reference within a document
type DocRef struct {
	documentId string
	token      string
	start      int
	end        int
}

// QueryDocument is the data structure returned by Index.Query()
type QueryDocument struct {
	documentId string
	refs       []WordRef
	tokens     []string
}

func (qd *QueryDocument) setTokens() {
	tokens := make([]string, 0, 10)
	tokenMap := make(map[string]bool, 10)
	for _, ref := range qd.refs {
		_, ok := tokenMap[ref.token]
		if !ok {
			tokens = append(tokens, ref.token)
			tokenMap[ref.token] = true
		}
	}
	qd.tokens = tokens
}

// WordRef references a word within a Document's text
// You may it for highlighting
type WordRef struct {
	token string
	start int
	end   int
}

// Helper struct for async search operations
type AsyncSearch struct {
	searchTerm string
	queryDocs  []QueryDocument
}

func (as *AsyncSearch) setQueryDocs(queryDocs []QueryDocument) {
	as.queryDocs = queryDocs
}

// Inverted Index
type TokenMap map[string][]DocRef

// Filter function signature
type filterFunc func(string) string

// Index
type Index struct {
	tokenMap *TokenMap
}

// Do query searchTerm concurrently within index
func (index *Index) QueryConcurrent(searchTerm string) []QueryDocument {
	// split searchTerm into words
	searchTokens := tokenize(searchTerm, func(value string) string {
		return value // no op filter
	})

	// Create an asyncSearch data for each searchTerm
	asyncSearches := make([]AsyncSearch, 0, len(searchTokens))
	for _, searchToken := range searchTokens {
		asyncSearches = append(asyncSearches,
			AsyncSearch{
				searchTerm: searchToken.value,
				queryDocs:  make([]QueryDocument, 0, 10)})
	}

	// execute search for each searchTerm as goroutine
	// write search result to asyncSearch.queryDocs
	var wg sync.WaitGroup
	wg.Add(len(asyncSearches))
	for i := 0; i < len(asyncSearches); i++ {
		go index.queryAsync(&asyncSearches[i], &wg)
	}
	wg.Wait()

	// merge results from asyncSearch.queryDocs
	queryDocs := make([]QueryDocument, 0, 10)
	for _, asyncSearch := range asyncSearches {
		queryDocs = merge(queryDocs, asyncSearch.queryDocs)
	}

	// rank / sort queryDocs
	rankQueryDocuments(queryDocs)

	return queryDocs
}

func (index *Index) queryAsync(asyncSearch *AsyncSearch, wg *sync.WaitGroup) {
	defer wg.Done()
	queryDocs := index.Query(asyncSearch.searchTerm)
	asyncSearch.setQueryDocs(queryDocs)
}

// Do query searchTerm within index
func (index *Index) Query(searchTerm string) []QueryDocument {
	// map text to searchTokens
	searchTokens := analyze(searchTerm, filters())

	// string builder for log output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Index search:'%s' ->", searchTerm))

	// store results to resultsRef
	resultDocRefs := make([]DocRef, 0, len(searchTokens))

	// remember handled tokens
	handledTokenMap := make(map[string]bool, len(searchTokens))

	// index lookup for every search token
	for _, searchToken := range searchTokens {
		if _, handled := handledTokenMap[searchToken.value]; handled {
			// ignore token, already handled
			continue
		}
		handledTokenMap[searchToken.value] = true

		// lookup token
		docRefs := (*index.tokenMap)[searchToken.value]
		if len(docRefs) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf(" %s/%d", searchToken.value, len(docRefs)))
		// add to result
		for _, docRef := range docRefs {

			resultDocRefs = append(resultDocRefs, docRef)
		}
	}
	log.Printf(sb.String())
	// transform to QueryDocuments
	queryDocs := queryDocument(resultDocRefs)

	rankQueryDocuments(queryDocs)

	return queryDocs
}

// Rank / Sort query documents by len(tokens) and then by len(refs):w
func rankQueryDocuments(queryDocs []QueryDocument) {

	// set tokens from wordRef
	for i := 0; i < len(queryDocs); i++ {
		queryDocs[i].setTokens()
	}

	// sort (rank)
	sort.Slice(queryDocs, func(i, j int) bool {
		switch diff := len(queryDocs[i].tokens) - len(queryDocs[j].tokens); {
		case diff > 0:
			return true
		case diff < 0:
			return false
		default:
			return len(queryDocs[i].refs) > len(queryDocs[j].refs)

		}
	})
}

func queryDocument(docRefs []DocRef) []QueryDocument {
	queryDocs := make(map[string]QueryDocument, 10)

	for _, docRef := range docRefs {
		documentId := docRef.documentId
		queryDoc, found := queryDocs[documentId]
		if !found {
			queryDoc = QueryDocument{documentId: docRef.documentId, refs: make([]WordRef, 0, 10)}
		}
		queryDoc.refs = append(queryDoc.refs,
			WordRef{start: docRef.start,
				end:   docRef.end,
				token: docRef.token})
		queryDocs[documentId] = queryDoc
	}
	docs := make([]QueryDocument, 0, len(queryDocs))
	for _, doc := range queryDocs {
		docs = append(docs, doc)
	}
	return docs
}

// crate Document inex
func CreateIndex(documents []Document) Index {
	log.Printf("Index init")
	filters := filters()
	log.Printf("Index analyze documents")
	tokenMap := make(TokenMap)
	for _, document := range documents {
		tokens := analyze(document.text, filters)
		documentTokenMap := documentTokenMap(document.id, tokens)
		tokenMap = mergeMaps(tokenMap, documentTokenMap)
	}

	log.Printf("Index ready")
	return Index{tokenMap: &tokenMap}
}

// add source entries to target map
func mergeMaps(m1 TokenMap, m2 TokenMap) TokenMap {
	mergeMap := make(map[string][]DocRef, len(m1)+len(m2))
	for k, v := range m1 {
		docRefs := make([]DocRef, 0, len(v))
		mergeMap[k] = append(docRefs, v...)
	}
	for k, v := range m2 {
		docRefs, found := mergeMap[k]
		if !found {
			docRefs = make([]DocRef, 0, len(v))
		}
		mergeMap[k] = append(docRefs, v...)
	}
	return mergeMap
}

func analyze(text string, filters filterFunc) []Token {
	return tokenize(text, filters)
}

var _filters filterFunc

func filters() filterFunc {
	if _filters != nil {
		// use cached filters
		return _filters
	}
	log.Printf("Index add filter lowercase")
	lowercase := lowerCaseFilter()
	log.Printf("Index add filter stopwords")
	stopword := stopwordFilter()
	log.Printf("Index add filter stemmer")
	stemmer := stemmerFilter()
	_filters = func(text string) string {
		text = lowercase(text)
		text = stopword(text)
		text = stemmer(text)
		return text
	}
	return _filters
}

func tokenize(text string, filter filterFunc) []Token {
	tokens := make([]Token, 0, len(text)/10)
	endOfText := len(text) - 1
	tokenStart := -1
	for i, c := range text {
		if !(unicode.IsLetter(c) || unicode.IsNumber(c)) || i == endOfText {
			tokenEnd := i
			if i == endOfText {
				tokenEnd++
			}
			if tokenEnd > tokenStart && tokenStart != -1 {
				token := text[tokenStart:tokenEnd]
				token = filter(token)
				if len(token) > 0 {
					tokens = append(tokens, Token{value: token, start: tokenStart, end: i})
				}
			}
			tokenStart = -1
		} else if tokenStart == -1 {
			tokenStart = i
		}
	}
	return tokens
}

func stemmerFilter() filterFunc {
	return func(value string) string {
		return snowballeng.Stem(value, false)
	}
}

func stopwordFilter() filterFunc {
	stopwords := readStopwords()
	return func(value string) string {
		if _, b := stopwords[value]; !b {
			return value
		} else {
			return ""
		}
	}
}

// read stopwords_en.txt file
func readStopwords() map[string]bool {
	stopwords_en, err := ioutil.ReadFile("stopwords_en.txt")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	stopwords_list := strings.Split(string(stopwords_en), "\n")
	stopwords := make(map[string]bool, len(stopwords_list))
	for _, word := range stopwords_list {
		stopwords[word] = true
	}
	return stopwords
}

func lowerCaseFilter() filterFunc {
	return func(value string) string {
		return strings.ToLower(value)
	}
}

func documentTokenMap(documentId string, tokens []Token) TokenMap {
	tokenMap := make(TokenMap, cap(tokens)/10)
	for _, token := range tokens {
		docRefs, found := tokenMap[token.value]
		if !found {
			docRefs = make([]DocRef, 0, 10)
		}
		docRefs = append(docRefs, DocRef{
			token:      token.value,
			documentId: documentId,
			start:      token.start,
			end:        token.end})
		tokenMap[token.value] = docRefs
	}
	return tokenMap
}

func merge(qds1 []QueryDocument, qds2 []QueryDocument) []QueryDocument {
	result := make([]QueryDocument, 0, len(qds1)+len(qds2))
	// append qds1
	for _, qd1 := range qds1 {
		queryDoc := QueryDocument{
			documentId: qd1.documentId,
			refs:       make([]WordRef, 0, len(qd1.refs))}
		queryDoc.refs = append(queryDoc.refs, qd1.refs...)
		result = append(result, queryDoc)
	}
	// append qds2
	for _, qd2 := range qds2 {
		existingIdx := findIndex(result, qd2.documentId)
		if existingIdx >= 0 {
			queryDoc := result[existingIdx]
			queryDoc.refs = append(queryDoc.refs, qd2.refs...)
			result[existingIdx] = queryDoc
		} else {
			queryDoc := QueryDocument{
				documentId: qd2.documentId,
				refs:       make([]WordRef, 0, len(qd2.refs))}
			queryDoc.refs = append(queryDoc.refs, qd2.refs...)
			result = append(result, queryDoc)
		}
	}
	return result
}

func findIndex(queryDocs []QueryDocument, documentId string) int {
	for i, qd := range queryDocs {
		if qd.documentId == documentId {
			return i
		}
	}
	return -1
}
