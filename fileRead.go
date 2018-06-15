package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"bufio"
)

func init() {

}

// File Study code
func main() {

	isFileExist()

	readAllFileBySeek()

	readALlFile()

	readLineByLine()
}

func isFileExist() bool {

	var filePath = "log.log"

	_, err := os.Stat(filePath)

	if err != nil {
		return false
	}

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func readAllFileBySeek() {

	filePath := "log.log"

	fd, err := os.Open(filePath)

	defer fd.Close()

	if err != nil {
		panic(err.Error())
	}

	fileLength, err := fd.Seek(0, 2)

	if err != nil {
		panic(err)
	}

	fileBytes := make([]byte, fileLength)

	fd.Read(fileBytes)

	fmt.Println(string(fileBytes))
}

func readALlFile() {

	filePath := "log.log"

	allFIleBytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(allFIleBytes))
}

func readLineByLine() {

	filePath := "log.log"

	fd, _ := os.Open(filePath)

	defer fd.Close()

	// 若指针已经产生位移则会从位移处产生开始读取
	// 若确定指针在开始处，则可忽略seek
	fd.Seek(0, 0)

	input := bufio.NewScanner(fd)

	num := 1

	for input.Scan() {
		fmt.Print(num, ": ")
		fmt.Println(input.Text())

		num++
	}
}
