package main

import (
	"net/http"
	"golang.org/x/net/websocket"
	"io"
	"time"
	"syscall"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"regexp"
	"strconv"
	"github.com/go-redis/redis"
)

func main() {

	setSocketServer()
	setHttpServer()

	stayDown()
}

// Init
func init() {

}

// Main Func Stay in running
func stayDown() {

	var exit chan bool

	<-exit
}

type cpuStat struct {
	User float32 `json:"user"`
	Sys  float32 `json:"sys"`
	Idle float32 `json:"idle"`
	All  float32 `json:"all"`
}

type memoryStat struct {
	Used   int `json:"used"`
	UnUsed int `json:"unUsed"`
	All    int `json:"all"`
}

type diskStat struct {
	All  uint64 `json:"all"`
	Free uint64 `json:"free"`
	Used uint64 `json:"used"`
}

// Get Stat from Linux
func getLinuxMonitor() (*cpuStat, *memoryStat) {

	output, err := exec.Command("top", `-b -n 1`).CombinedOutput()

	if err != nil {
		panic(err.Error())
	}

	stringMap := strings.Split(string(output), "\n")
	stringMap = stringMap[:5]

	linuxCpuStat := new(cpuStat)
	linuxMomoryStat := new(memoryStat)

	for _, value := range stringMap {

		if strings.Contains(value, "Cpu") {

			reg := regexp.MustCompile(`(\d+\.\d+)`)

			matches := reg.FindAllString(value, 4)

			newUsed, _ := strconv.ParseFloat(matches[0], 32)
			newIdle, _ := strconv.ParseFloat(matches[3], 32)
			newSys, _ := strconv.ParseFloat(matches[1], 32)

			linuxCpuStat.Sys = float32(newSys)
			linuxCpuStat.User = float32(newUsed)
			linuxCpuStat.Idle = float32(newIdle)
			linuxCpuStat.All = linuxCpuStat.Sys + linuxCpuStat.User + linuxCpuStat.Idle
		}

		if strings.Contains(value, "Mem") {

			reg := regexp.MustCompile(`(\d+)`)

			matches := reg.FindAllString(value, 2)

			linuxMomoryStat.All, _ = strconv.Atoi(matches[0])
			linuxMomoryStat.Used, _ = strconv.Atoi(matches[1])
			linuxMomoryStat.UnUsed, _ = strconv.Atoi(matches[2])
		}
	}

	return linuxCpuStat, linuxMomoryStat
}

// Get Stat From Mac
func getMacOsMonitor() (*cpuStat, *memoryStat) {

	output, err := exec.Command("top", `-l 1`).CombinedOutput()

	if err != nil {
		panic(err.Error())
	}

	stringMap := strings.Split(string(output), "\n")
	stringMap = stringMap[:10]

	macOsCpuStat := new(cpuStat)
	macOsMemoryStat := new(memoryStat)

	for _, value := range stringMap {

		if strings.Contains(value, "CPU") {

			reg := regexp.MustCompile(`(\d+\.\d+)`)

			matches := reg.FindAllString(value, -1)

			userValue, _ := strconv.ParseFloat(matches[0], 32)
			sysValue, _ := strconv.ParseFloat(matches[1], 32)
			idleValue, _ := strconv.ParseFloat(matches[2], 32)

			macOsCpuStat.User = float32(userValue)
			macOsCpuStat.Sys = float32(sysValue)
			macOsCpuStat.Idle = float32(idleValue)
			macOsCpuStat.Idle = 100
		}

		if strings.Contains(value, "PhysMem") {

			reg := regexp.MustCompile(`(\d+)`)

			match := reg.FindAllString(value, -1)

			macOsMemoryStat.Used, _ = strconv.Atoi(match[0])
			macOsMemoryStat.UnUsed, _ = strconv.Atoi(match[2])
			macOsMemoryStat.All = macOsMemoryStat.UnUsed + macOsMemoryStat.Used
		}
	}

	return macOsCpuStat, macOsMemoryStat
}

// Get Cpu And Memory stat
func getStatMonitor() (*cpuStat, *memoryStat) {

	switch runtime.GOOS {
	case "linux":
		return getLinuxMonitor()
	case "darwin":
		return getMacOsMonitor()
	default:
		panic("System Do Not Support")
	}
}

// Get Disk Stat
func getDiskMonitor() *diskStat {

	data := new(diskStat)
	fs := syscall.Statfs_t{}

	err := syscall.Statfs("/", &fs)

	if err != nil {
		return data
	}

	data.All = fs.Blocks * uint64(fs.Bsize) / 1024 / 1024 / 1024
	data.Free = fs.Bfree * uint64(fs.Bsize) / 1024 / 1024 / 1024
	data.Used = data.All - data.Free

	return data
}

// Add http ping
func setHttpServer() {

	http.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, "ok")
	})

	err := http.ListenAndServe(":9998", nil)

	if err != nil {
		panic(err.Error())
	}
}

// Set socket server
func setSocketServer() {

	http.Handle("/sub", websocket.Handler(func(ws *websocket.Conn) {

		defer ws.Close()

		msg := make([]byte, 512)

		_, err := ws.Read(msg)

		if err != nil {
			ws.Close()
		}

		timer(ws)
	}))

	err := http.ListenAndServe(":9999", nil)

	if err != nil {
		panic(err.Error())
	}

	http.Handle("/", http.FileServer(http.Dir(".")))
}

// Send Data to every Conn by Timer
func timer(ws *websocket.Conn) {

	timer := time.NewTicker(1 * time.Second)

	for {

		select {

		case <-timer.C:

			writeString := monitorMate()

			_, err := ws.Write([]byte(writeString))

			if err != nil {
				ws.Close()
			}
		}
	}

	ws.Close()
}

// Output Data Mate
func monitorMate() string {

	data := struct {
		DiskStat   *diskStat         `json:"diskStat"`
		CpuStat    *cpuStat          `json:"cpuStat"`
		MemoryStat *memoryStat       `json:"memoryStat"`
		RedisStat  map[string]string `json:"redisStat"`
	}{}

	data.DiskStat = getDiskMonitor()
	data.RedisStat = getRedisMonitor()
	data.CpuStat, data.MemoryStat = getStatMonitor()

	result, _ := json.Marshal(data)

	return string(result)
}

func getRedisMonitor() map[string]string {

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := redisClient.Ping().Result()

	stringMap := make(map[string]string)

	if err != nil {
		return stringMap
	}

	output := redisClient.Info("Memory")
	infoMap := strings.Split(output.Val(), "\r\n")
	infoMap = infoMap[1:] // remove #Memory

	for _, value := range infoMap {

		valueSlice := strings.Split(value, ":")

		if len(valueSlice) == 2 {
			stringMap[valueSlice[0]] = valueSlice[1]
		}
	}

	return stringMap
}
