package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
)

func main() {
	pid := getProcessPid()

	writePid(pid)

	newPid := readPid()

	fmt.Printf("pid is %d\n", newPid)
}

func getProcessPid() int {
	return os.Getpid()
}

func writePid(pid int) {

	fileName := "./pid"

	// create file permissions is 644
	err := ioutil.WriteFile(fileName, []byte(strconv.Itoa(pid)),  0644)

	if err != nil {
		panic(err)
	}
}

func readPid() int {

	fileName := "./pid"

	pid, err := ioutil.ReadFile(fileName)

	if err != nil {
		panic(err)
	}

	pidInt, err := strconv.Atoi(string(pid))

	return pidInt
}
