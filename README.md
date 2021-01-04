# ShakeSearch

Welcome to the Pulley Shakesearch Take-home Challenge! In this repository,
you'll find a simple web app that allows a user to search for a text string in
the complete works of Shakespeare.

You can see a live version of the app at
https://pulley-shakesearch.herokuapp.com/. Try searching for "Hamlet" to display
a list of results.

In it's current state, however, the app is just a rough prototype. The search is
case sensitive, the results are difficult to read, and the search is limited to
exact matches.

## Improvements 

You can see a live version of the improved ap at https://ts-shakesearch.herokuapp.com/.

Try searching for "Nile gods the" to display s set of results.

You will see the following improvements:
* Ranked result list containing references to the search terms in Shakespeare's works
* Stopwords like 'the' are ignored
* Multiple forms of the word 'god' are found (God's, God, gods)

### Document Parser

[documents.go](./documents.go) does:

* Read completeworks.txt
* Find work start and end markers
* Create a document for each of Shakespeare's works

### Indexer

[Index](./index.go) implements an inverted map based alogorithm for fast full-text search. The index holds a list of word references for each token. 

Documents are indexed on startup of the application. The Index then provides two different query methods: 
* Query(searchTerm string) []QueryDocument
* QueryConcurrent(searchTerm string) []QueryDocument

The concurrent query method splits the search term to multiple tokens. A concurrent lookup for each token is executed in different threads using goroutines.

The Index analyses documents as follows:
1. Tokenize documents (extract words)
2. Apply filters to each token
   - Lower case filter
   - Stopword filter (remove common words, see stopwords_en.txt)
   - Apply stemmer (normalize forms of the same word, e.g. fish, fishes -> fish)

The index stores the output of the analyzer. The data structure is defined as follows:
* Inverted Map: map[token] [] DocRef(*document, start, end) ]
* Reference to a word withn a document: DocRef(*document, start, end)

A query against the index works as follows:
1. Analyze query term to get search tokens
2. Lookup index for all analyzed tokens
3. Sort (rank) search results

### Further Improvements

* Support fuzzy search using a Radix Tree 
* Fix document parser encoding

## Your Mission

Improve the search backend. Think about the problem from the user's perspective
and prioritize your changes according to what you think is most useful.

## Submission

1. Fork this repository and send us a link to your fork after pushing your changes. 
2. Heroku hosting - The project includes a Heroku Procfile and, in its
current state, can be deployed easily on Heroku's free tier.
3. In your submission, share with us what changes you made and how you would prioritize changes if you had more time.


