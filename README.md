# ShakeSearch

Welcome to the Pulley Shakesearch Take-home Challenge! In this repository,
you'll find a simple web app that allows a user to search for a text string in
the complete works of Shakespeare.

You can see a live version of the app at
https://pulley-shakesearch.herokuapp.com/. Try searching for "Hamlet" to display
a set of results.

In it's current state, however, the app is just a rough prototype. The search is
case sensitive, the results are difficult to read, and the search is limited to
exact matches.

## Submission Improvements 

You can see a live version of the improved ap at https://ts-shakesearch.herokuapp.com/.

Try searching for "Nile gods the" to display s set of results.

You will see the following improvements:
* a ranked result list containing references to the search terms in Shakespeare's works
* stopwords like 'the' are ignored
* multiple forms of the word 'god' are found

### Document Parser

[documents.go](./documents.go) does:

* read completeworks.txt
* find work start / end markers
* create a document for each of Shakespeare's works

### Indexer

Index is an inverted map based alogorithm for fast full-text search based on tokens.

The Index analyses documents as follows:
1. tokenize documents (extract words)
2. apply filters to each token
2.1. lower case filter
2.2. stopword filter (remove common words, see stopwords_en.txt)
2.3. apply stemmer (normalize forms of the same word, e.g. fish, fishes -> fish)

The index stores the output of the analyzer. The data structure is defined as follows:
* Inverted Map: map[token] [] DocRef(*document, start, end) ]
* Reference to a word withn a document: DocRef(*document, start, end)

A query against the index works as follows:
1. analyze query term to get search tokens
2. lookup index for all analyzed tokens
3. sort (rank) search results


## Your Mission

Improve the search backend. Think about the problem from the user's perspective
and prioritize your changes according to what you think is most useful.

## Submission

1. Fork this repository and send us a link to your fork after pushing your changes. 
2. Heroku hosting - The project includes a Heroku Procfile and, in its
current state, can be deployed easily on Heroku's free tier.
3. In your submission, share with us what changes you made and how you would prioritize changes if you had more time.


