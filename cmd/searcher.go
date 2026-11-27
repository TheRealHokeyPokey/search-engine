package main

import (
	"bufio"
	"cmp"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"maps"
	"math"
	"myproject/common"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

// Returns the inverse document frequency of a query term.
func getQueryIdf(thisPostings map[string]int64, nDocuments int64) float64 {
	nqi := int64(len(thisPostings))
	return math.Log((float64(nDocuments-nqi)+0.5)/(float64(nqi)+0.5) + 1)
}

// Computes and prints the Okapi BM25 scores of relevant documents.
// BM25 formula is from Wikipedia: https://en.wikipedia.org/wiki/Okapi_BM25
func printRelevantBM25Scores(queryWords []string,
	dictionary map[string]common.DictEntry,
	avgDocumentLength float64,
	nDocuments int64,
	postingsFile *os.File,
	docToLength map[string]int64,
) {
	docScores := make(map[string]float64)
	candidateDocs := make(map[string]bool)
	queryIdfs := make([]float64, 0)
	queryPostingsSlice := make([]map[string]int64, 0)

	// values from lecture 3
	// note: Wikipedia formula did not use k3
	k1 := 1.2
	b := 0.75

	// collect queryPostingses, queryIdfs and docNos of relevant documents
	for i := range queryWords {
		queryPostings := getQueryPostings(queryWords[i], dictionary, postingsFile)
		queryPostingsSlice = append(queryPostingsSlice, queryPostings)
		qIdf := getQueryIdf(queryPostings, nDocuments)
		for key := range queryPostings {
			candidateDocs[key] = true
		}
		queryIdfs = append(queryIdfs, qIdf)
	}
	// compute BM25 scores for each relevant document
	for doc := range candidateDocs {
		docScore := 0.0
		docLength := docToLength[doc]
		docLengthFloat := float64(docLength)
		for i := range queryWords {
			idf := queryIdfs[i]
			fQiD := float64(queryPostingsSlice[i][doc])
			term := (fQiD * (k1 + 1.0)) / (fQiD + k1*(1.0-b+b*(docLengthFloat/avgDocumentLength)))
			docScore += idf * term
		}
		docScores[doc] = docScore
	}

	// sort relevant documents by score and write to stdout
	sortedKeys := slices.SortedFunc(maps.Keys(docScores), func(a, b string) int {
		return cmp.Compare(docScores[b], docScores[a])
	})
	stdoutWriter := bufio.NewWriter(os.Stdout)
	for _, key := range sortedKeys {
		fmt.Fprintf(stdoutWriter, "%s %f\n", key, docScores[key])
	}
	stdoutWriter.Flush()
}

// Quickly reads the postings for a query term from an index file.
// Returns a map from docNo to frequency of query in document for each
// document that contains the query term.
func getQueryPostings(query string, dictionary map[string]common.DictEntry, postingsFile *os.File) map[string]int64 {
	dictEntry, _ := dictionary[query]
	location := dictEntry.Loc
	length := dictEntry.Len
	_, _ = postingsFile.Seek(location, 0)
	indexReader := bufio.NewReader(postingsFile)
	b := make([]byte, length)
	_, _ = io.ReadFull(indexReader, b)
	resultString := string(b)

	queryPostings := make(map[string]int64)
	for line := range strings.Lines(resultString) {
		lineSlice := strings.Split(line, ",")
		docNo := lineSlice[0]
		docFreq, _ := strconv.ParseInt(strings.TrimSpace(lineSlice[1]), 10, 64)
		queryPostings[docNo] = docFreq
	}
	return queryPostings
}

func main() {
	wd, _ := os.Getwd()

	dictPath := path.Join(wd, "data", "wsj-dict.gob")
	dictF, _ := os.Open(dictPath)
	defer common.CloseFile(dictF)
	// dictionary maps query terms to (loc, len) which describe
	// where the postings are located in the index file.
	dictionary := make(map[string]common.DictEntry)
	d := gob.NewDecoder(bufio.NewReader(dictF))
	err := d.Decode(&dictionary)
	common.Check(err)

	// index is the postings. For a query term it contains the docNo
	// of all documents that contain the term and the frequency of the
	// term within that document. We do not read the whole index into memory,
	// only the parts that are required.
	indexPath := path.Join(wd, "data", "wsj-index.txt")
	indexF, _ := os.Open(indexPath)
	defer common.CloseFile(indexF)

	docToLengthPath := path.Join(wd, "data", "wsj-doctolen.gob")
	docToLengthF, _ := os.Open(docToLengthPath)
	defer common.CloseFile(docToLengthF)
	// docToLength maps docNo to the length of the document in words.
	docToLength := make(map[string]int64)
	d = gob.NewDecoder(bufio.NewReader(docToLengthF))
	err = d.Decode(&docToLength)
	common.Check(err)

	var nDocuments int64
	indexReader := bufio.NewReader(indexF)
	binary.Read(indexReader, binary.LittleEndian, &nDocuments)

	var averageDocumentLength float64
	binary.Read(indexReader, binary.LittleEndian, &averageDocumentLength)

	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
	scanner.Split(bufio.ScanWords)
	var queryTerms []string
	for scanner.Scan() {
		term := scanner.Text()
		term = strings.ToLower(term)
		queryTerms = append(queryTerms, term)
	}
	printRelevantBM25Scores(queryTerms, dictionary, averageDocumentLength, nDocuments, indexF, docToLength)
}
