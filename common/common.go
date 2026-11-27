package common

import "os"

const (
	STARTTAGLENGTH = len("<DOCNO>")
	ENDTAGLENGTH   = len("</DOCNO>")
)

func CharIsAlphaNumeric(r byte) bool {
	return (r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
}

type DictEntry struct {
	Loc int64
	Len int64
}

func CloseFile(f *os.File) {
	err := f.Close()
	if err != nil {
		panic(err)
	}
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
