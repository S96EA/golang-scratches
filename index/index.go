package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/icrowley/fake"
)

var fileName = ""
var indexFileName = ""

type Row struct {
	id   int32
	name string
	text string
}

type Db struct {
	file  *os.File
	index map[int64]int64
}

func NewDb() *Db {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	db := &Db{}
	db.file = file
	db.index = readIndex()
	return db
}

// index data
func indexData() {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	indexFile, err := os.OpenFile(indexFileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer indexFile.Close()

	for i := 0; i < 80000; i++ {
		var totals int32
		if err := binary.Read(file, binary.LittleEndian, &totals); err != nil { // read totals
			panic(err)
		}
		var id int32
		if err := binary.Read(file, binary.LittleEndian, &id); err != nil { // read id
			panic(err)
		}
		offset, err := file.Seek(-4, io.SeekCurrent) // id offset
		if err != nil {
			panic(err)
		}
		if _, err := file.Seek(int64(totals), io.SeekCurrent); err != nil {
			panic(err)
		}
		if _, err := indexFile.WriteString(fmt.Sprintf("%v,%v\n", id, offset)); err != nil {
			panic(err)
		}
	}
}

func readIndex() map[int64]int64 {
	indexFile, err := os.OpenFile(indexFileName, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer indexFile.Close()
	reader := bufio.NewReader(indexFile)
	idToOffset := make(map[int64]int64)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lineData := strings.Split(strings.TrimSpace(line), ",")
		id, _ := strconv.ParseInt(lineData[0], 10, 64)
		offset, err := strconv.ParseInt(lineData[1], 10, 64)
		idToOffset[id] = offset
	}
	return idToOffset
}

func (db *Db) readDataByIndex(id int64) (*Row, error) {
	offset, ok := db.index[id]
	if !ok {
		return nil, errors.New("id not found")
	}
	_, err := db.file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}
	var id32 int32
	if err := binary.Read(db.file, binary.LittleEndian, &id32); err != nil {
		return nil, err
	}
	if int64(id32) != id {
		return nil, errors.New("id not equals")
	}
	var nameLen int32
	if err := binary.Read(db.file, binary.LittleEndian, &nameLen); err != nil {
		return nil, err
	}
	nameBytes := make([]byte, nameLen)
	if _, err := db.file.Read(nameBytes); err != nil {
		return nil, err
	}
	var textLen int32
	if err := binary.Read(db.file, binary.LittleEndian, &textLen); err != nil {
		return nil, err
	}
	textBytes := make([]byte, textLen)
	if _, err := db.file.Read(textBytes); err != nil {
		return nil, err
	}
	return &Row{
		id:   id32,
		name: string(nameBytes),
		text: string(textBytes),
	}, nil
}

func main() {
	db := NewDb()
	for {
		var id int64
		fmt.Scanln(&id)
		start := time.Now()
		row, err := db.readDataByIndex(id)
		if err != nil {
			panic(err)
		}
		//fmt.Println(row.id)
		fmt.Println(row.name)
		//fmt.Println(len(row.text))
		fmt.Println(time.Now().Sub(start).Milliseconds())
	}
}

func readData() {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	start := time.Now()
	for i := 0; i < 80000; i++ {
		var totals int32
		if err := binary.Read(file, binary.LittleEndian, &totals); err != nil { // read totals
			panic(err)
		}
		var id int32
		if err := binary.Read(file, binary.LittleEndian, &id); err != nil { // read id
			panic(err)
		}
		var nameLen int32
		if err := binary.Read(file, binary.LittleEndian, &nameLen); err != nil {
			panic(err)
		}
		var nameBytes []byte = make([]byte, nameLen)
		if n, err := file.Read(nameBytes); int32(n) != nameLen || err != nil {
			panic(err)
		}
		//name := string(nameBytes)
		var textLen int32
		if err := binary.Read(file, binary.LittleEndian, &textLen); err != nil {
			panic(err)
		}
		var textBytes []byte = make([]byte, textLen)
		if n, err := file.Read(textBytes); int32(n) != textLen || err != nil {
			panic(err)
		}
		//text := string(textBytes)
		println(id)
	}
	fmt.Println(time.Now().Sub(start).Milliseconds(), "ms")
}

func genData() {
	rand.Seed(time.Now().UnixNano())
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i := 0; i < 80000; i++ {
		name := fake.FullName()
		text := name + ":" + fake.CharactersN(int(rand.Int31n(100000)))
		totals := len(name) + len(text) + 4*3
		if err := binary.Write(file, binary.LittleEndian, int32(totals)); err != nil {
			panic(err)
		}
		if err := binary.Write(file, binary.LittleEndian, int32(i)); err != nil { // id
			panic(err)
		}
		if err := binary.Write(file, binary.LittleEndian, int32(len(name))); err != nil { // len(name)
			panic(err)
		}
		if _, err := file.WriteString(name); err != nil {
			panic(err)
		}
		if err := binary.Write(file, binary.LittleEndian, int32(len(text))); err != nil { // len(text)
			panic(err)
		}
		if _, err := file.WriteString(text); err != nil {
			panic(err)
		}
	}
}
