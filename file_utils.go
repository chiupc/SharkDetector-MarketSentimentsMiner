package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var fileMutex sync.Mutex

func writeToCSV(ctx context.Context,filename string, linesToWrite []string){
	fileMutex.Lock()
	log := logger.WithFields(logrus.Fields{
		"ctx-id": ctx.Value("ctx-id"),
	})
	log.Debug(fmt.Sprintf("writing total number of lines: %d",len(linesToWrite)))
	//var file io.Writer
	//if _, err := os.Stat(filepath.Join("tmp",filename)); err == nil {
	//	// path/to/whatever exists
	//	file = os.OpenFile()
	//} else if os.IsNotExist(err) {
	//	file, err = os.Create(fmt.Sprintf("%s.csv",filename))
	//	if err != nil {
	//		log.Error(err)
	//	}
	//}
	file, err := os.OpenFile(filepath.Join("tmp",filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err)
	}else {
		writer := bufio.NewWriterSize(file, MaxBufferSize)
		for i, line := range linesToWrite {
			log.Debug(fmt.Sprintf("line number: %d, text: %s", i, line))
			bytesWritten, err := writer.WriteString(line + "\n")
			if err != nil {
				log.Error(fmt.Sprintf("Got error while writing to a file. Err: %s", err.Error()))
			}
			log.Debug(fmt.Sprintf("Bytes Written: %d\n", bytesWritten))
			log.Debug(fmt.Sprintf("Available: %d\n", writer.Available()))
			log.Debug(fmt.Sprintf("Buffered : %d\n", writer.Buffered()))
		}
		writer.Flush()
		defer file.Close(); fileMutex.Unlock()
	}
}

func generateFileName(ctx context.Context, quote string, startTime int, endTime int) string{
	filename := strings.Join([]string{quote,strconv.Itoa(int(startTime)),strconv.Itoa(int(endTime)), strings.Split(ctx.Value("ctx-id").(string),"-")[4]},"_")
	return filename
}

func cleanupYfConversation(s string) string{
	//remove new line
	s = strings.Replace(s, "\n", " ", -1)
	return s
}