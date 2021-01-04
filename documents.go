package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Parse completeworks.txt and create a Document for each of Shakespeare's works

// Document represents w single work out of Shakespeares's collection.
type Document struct {
	id    string // id of the document
	title string // title of work
	text  string // text of work, without title
}

// _DocMarker is a Helper struct. The Parser uses _DocMarker to collect information about
// documents in completeworks.txt
type _DocMarker struct {
	title     string // title of work
	start     int    // start of work, including title
	startText int    // start of works text, without title
}

type DocumentStore struct {
	documents     []Document
	documentsById map[string]Document
}

func (store *DocumentStore) all() []Document {
	return store.documents
}
func (store *DocumentStore) documentById(id string) (Document, error) {
	document, ok := store.documentsById[id]
	if !ok {
		return document, errors.New(fmt.Sprintf("Document with id '%s' not found.", id))
	}
	return document, nil
}

func LoadDocumentStore() DocumentStore {
	documents := loadDocuments()
	documentsById := make(map[string]Document, len(documents))
	for i := 0; i < len(documents); i++ {
		documentsById[documents[i].id] = documents[i]
	}
	return DocumentStore{documents: documents, documentsById: documentsById}
}

// Parse completeworks.txt and return Documents
func loadDocuments() []Document {
	log.Printf("Parser parse documents")
	text := readFile()

	// parse text to get markers (title + start of text)
	markers := parse(text)

	// create documents from markers and text
	documents := documents(markers, text)

	return documents
}

// Transform _DocMarkers to Documents
func documents(markers []_DocMarker, text string) []Document {
	documents := make([]Document, 0, len(markers))
	for i, marker := range markers {
		var endIndex int
		if i == len(markers)-1 {
			endIndex = len(text)
		} else {
			endIndex = markers[i+1].start - 2
		}
		documents = append(documents,
			Document{
				id:    strconv.Itoa(i + 1),
				title: marker.title,
				text:  text[marker.startText:endIndex]})
	}

	return documents
}

// Parse text and return _DocMarkers
func parse(text string) []_DocMarker {
	lines := strings.Split(text, "\r\n")
	re := regexp.MustCompile(`(?i)((contents)|(.*dramatis.person.*))$`)
	prevText := ""
	markers := make([]_DocMarker, 0, 45)
	pos := 0

	for i, line := range lines {
		nextLine := len(line) + 2
		if i < 133 {
			// skip index
			pos += nextLine
			continue
		}
		lineTrimmed := strings.TrimSpace(line)
		if re.Match([]byte(lineTrimmed)) && isUpper(prevText) {
			// Content start found
			markers = append(markers,
				_DocMarker{
					title:     prevText,
					start:     pos,
					startText: pos + nextLine})

		} else if line == "THE SONNETS" ||
			line == "A LOVER’S COMPLAINT" ||
			line == "THE PASSIONATE PILGRIM" ||
			line == "THE PHOENIX AND THE TURTLE" ||
			line == "THE RAPE OF LUCRECE" ||
			line == "VENUS AND ADONIS" { // could use a map here

			// Content start found
			markers = append(markers,
				_DocMarker{
					title:     lineTrimmed,
					start:     pos,
					startText: pos + nextLine})

		} else if len(line) > 0 {
			// remember line - might be title of next content block
			prevText = lineTrimmed
		}
		// set pos to next line
		pos += nextLine
	}

	return markers
}

// Read Content
func readFile() string {
	fileName := "completeworks.txt"
	text, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return string(text)
}

// Is Upper Case string^
func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
