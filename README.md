# Search engine

This repository holds the source code for a search engine I wrote for a
university paper. Unfortunately, I do not have the rights to publish the data
we used.

## Broad architecture overview

This project consists of an indexer and a searcher. The indexer builds an
inverted index from a malformed XML dataset. The searcher takes queries and uses
the index to return a ranked list of relevant documents and their relevance scores.

### Indexer

Parses words from the data (which is slightly malformed XML) and writes a
dictionary and an index to disk. The dictionary maps words to (loc, len) pairs
which describe the relevant ranges of the index. The index holds document IDs and
the number of times a word occurs in that document.

```bash
# compile indexer
go build -o ./bin/indexer ./cmd/indexer.go
# run indexer (this expects a specific dataset at ./data/wsj.xml, or it won't work)
./bin/indexer
```

### Searcher

Reads the dictionary from disk and takes a query from stdin. Ranks each document
containing at least one word from the query using BM25. Write the list of ranked
documents to stdout.

```bash
# compile searcher
go build -o ./bin/searcher ./cmd/searcher.go
# run searcher
./bin/searcher < path/to/query
```
