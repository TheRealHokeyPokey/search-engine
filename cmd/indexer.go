package main

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"os"
	"path"
	"strconv"
	"strings"

	"myproject/common"
)

// Add word to postings (in memory, not to disk)
func addWordToIndex(dictionary map[string]int, postings *[]map[string]int, word string, docNo string) {
	postIndex, ok := dictionary[word]
	if !ok {
		*postings = append(*postings, make(map[string]int))
		postIndex = len(*postings) - 1
		dictionary[word] = postIndex
	}
	postEntry := (*postings)[postIndex]
	postEntry[docNo]++
}

func main() {
	wd, _ := os.Getwd()
	p := path.Join(wd, "data", "wsj.xml")
	xmlData, _ := os.ReadFile(p)
	xmlStringData := string(xmlData)
	var wordBuilder strings.Builder

	// dictionary is a word-> postings map
	dictionary := make(map[string]int)
	// postings is a slice of docno->freq maps
	postings := make([]map[string]int, 0)
	docLengths := make(map[string]int64)

	var docNo string
	var numWordsInDoc int
	for i := 0; i < len(xmlStringData); i++ {
		character := xmlStringData[i]
		if !common.CharIsAlphaNumeric(character) {
			if wordBuilder.Len() != 0 {
				lowerWord := strings.ToLower(wordBuilder.String())
				addWordToIndex(dictionary, &postings, lowerWord, docNo)
				wordBuilder.Reset()
				numWordsInDoc++
			}
			if character == '<' {
				var skip int
				if xmlStringData[i:i+common.STARTTAGLENGTH] == "<DOCNO>" {
					if docNo != "" {
						docLengths[docNo] = int64(numWordsInDoc)
					}
					numWordsInDoc = 0
					closeTagIndex := strings.Index(xmlStringData[i:], "</DOCNO>")
					docNo = strings.TrimSpace(xmlStringData[i+common.STARTTAGLENGTH : i+closeTagIndex])
					skip = closeTagIndex + common.ENDTAGLENGTH
				} else {
					skip = strings.IndexRune(xmlStringData[i:], '>')
				}
				i += skip
			}
		} else {
			wordBuilder.WriteByte(character)
		}
	}
	if wordBuilder.Len() != 0 {
		lowerWord := strings.ToLower(wordBuilder.String())
		addWordToIndex(dictionary, &postings, lowerWord, docNo)
		wordBuilder.Reset()
	}

	// write data to disk

	wd, _ = os.Getwd()
	indexFilePath := path.Join(wd, "data", "wsj-index.txt")
	indexF, _ := os.Create(indexFilePath)
	defer common.CloseFile(indexF)
	indexWriter := bufio.NewWriter(indexF)

	// newDict will be saved to disk. It will map query terms to (loc, len) pairs
	// which describe the location of the postings for that term, within the
	// index file.
	newDict := make(map[string]common.DictEntry)

	totalBytes := 0

	var nDocuments int64 = int64(len(docLengths))
	totalWordsInDocuments := int64(0)
	for doc := range docLengths {
		totalWordsInDocuments += docLengths[doc]
	}
	averageDocLength := float64(totalWordsInDocuments) / float64(nDocuments)

	// write number of documents as int64
	binary.Write(indexWriter, binary.LittleEndian, nDocuments)
	totalBytes += 8
	// write average doc length as float64
	binary.Write(indexWriter, binary.LittleEndian, averageDocLength)
	totalBytes += 8

	for word := range dictionary {
		totalBytesStart := totalBytes
		postIndex := dictionary[word]
		wordPostings := postings[postIndex]
		for docNo := range wordPostings {
			nb, err := indexWriter.WriteString(docNo)
			totalBytes += nb
			common.Check(err)
			nb, err = indexWriter.WriteString(",")
			totalBytes += nb
			common.Check(err)
			nb, err = indexWriter.WriteString(strconv.Itoa(wordPostings[docNo]))
			totalBytes += nb
			common.Check(err)
			nb, err = indexWriter.WriteString("\n")
			totalBytes += nb
			common.Check(err)
		}
		wordBytes := totalBytes - totalBytesStart
		newDict[word] = common.DictEntry{Loc: int64(totalBytesStart), Len: int64(wordBytes)}
	}
	// write dictionary to disk in binary
	dictFilePath := path.Join(wd, "data", "wsj-dict.gob")
	dictF, _ := os.Create(dictFilePath)
	defer common.CloseFile(dictF)
	dictWriter := bufio.NewWriter(dictF)
	e := gob.NewEncoder(dictWriter)
	_ = e.Encode(newDict)

	// write doc-length hashmap to disk in binary
	docToLengthPath := path.Join(wd, "data", "wsj-doctolen.gob")
	docToLengthF, _ := os.Create(docToLengthPath)
	defer common.CloseFile(docToLengthF)
	docToLengthWriter := bufio.NewWriter(docToLengthF)
	e = gob.NewEncoder(docToLengthWriter)
	_ = e.Encode(docLengths)
}
