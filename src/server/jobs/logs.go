package jobs

import (
	"bufio"
	"io"
	"log"
	"sync"
	"time"
)

type LogEntry struct {
	Line      string
	Timestamp time.Time
	Stdout    bool
}

type logs struct {
	sync.Mutex
	cond         *sync.Cond
	readingCoros int
	logs         []LogEntry
	jobID        JobID
}

func (logs *logs) read(pipe io.ReadCloser, stdout bool) {
	reader := bufio.NewReader(pipe)
	name := "stderr"
	if stdout {
		name = "stdout"
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("pipe", name, "of job", logs.jobID, "got closed")
			break
		}
		log.Println("Job", logs.jobID, "got log on", name, ":", line)
		entry := LogEntry{Line: line, Timestamp: time.Now(), Stdout: stdout}
		logs.append(entry)
	}
	logs.Lock()
	defer logs.Unlock()
	logs.readingCoros -= 1
	logs.cond.Broadcast()
}

func (logs *logs) append(entry LogEntry) {
	logs.Lock()
	defer logs.Unlock()
	logs.logs = append(logs.logs, entry)
	logs.cond.Broadcast()
}

// Returning 0 length indicates that there are no more logs to return
func (logs *logs) get(start, maxCount int) []LogEntry {
	logs.Lock()
	defer logs.Unlock()
	for start >= len(logs.logs) && logs.readingCoros > 0 {
		logs.cond.Wait()
	}
	return logs.logs[start:min(start+maxCount, len(logs.logs))]
}

func (logs *logs) size() int {
	logs.Lock()
	defer logs.Unlock()
	return len(logs.logs)
}

func newLogs(stdout, stderr io.ReadCloser, jobID JobID) *logs {
	result := &logs{readingCoros: 2, jobID: jobID}
	result.cond = sync.NewCond(result)
	go result.read(stdout, true)
	go result.read(stderr, false)
	return result
}
