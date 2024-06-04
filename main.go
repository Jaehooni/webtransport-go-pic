package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	
    wt := webtransport.Server{
        H3: http3.Server{
            Addr: ":4433",
        },
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }

	totalBytes := 0
	startTime := time.Now()
	fileSize := 0

	var delays []time.Duration
	var lastTime time.Time

    http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Connection established!")
        session, err := wt.Upgrade(w, r)
        if (err != nil) {
            log.Printf("Failed to upgrade: %v", err)
            return
        }

		// logFile, err := os.OpenFile("transmission_log.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0777)
		// if err != nil {
		// 	log.Fatalf("Failed to open log file: %v", err)
		// }
		// defer logFile.Close()

		// fmt.Fprintln(logFile, "Time,Bytes")

        go func() {
            stream, err := session.AcceptStream(context.Background())
            if (err != nil) {
                log.Printf("Failed to accept stream: %v", err)
                return
            }

            files, err := getFiles("images-folder")
            separator := []byte{0, 255, 0, 255}  // 더 명확한 구분자 사용
            for _, file := range files {
                if err := sendFile(file, stream, &fileSize); err != nil {
                    log.Printf("Error sending file %s: %v", file, err)
                    return
                }
                // 이미지 구분자로 null 바이트 전송
                if _, err := stream.Write(separator); err != nil {
                    log.Printf("Failed to write separator for %s: %v", file, err)
                    return
                }
				totalBytes += fileSize;

				// currentTime := time.Now()
				//fmt.Printf("%d %d\n", currentTime, fileSize);
				// fmt.Fprintf(logFile, "%s,%d\n", currentTime.Format(time.RFC3339Nano), fileSize)
				// fmt.fl
				// if err != nil {
				// 	log.Fatalf("Failed to open log file: %v", err)
				// }
				// defer logFile.Close()
				time.Sleep(100 * time.Millisecond)


				if !lastTime.IsZero(){
					delay := time.Since(lastTime) - 100 * time.Millisecond
					// fmt.Printf("delay: %v", delay);
					delays = append(delays, delay)
				}
				lastTime = time.Now()
				time.Sleep(100 * time.Millisecond)
            }

            // 모든 파일 전송 후 스트림 닫기
            if err := stream.Close(); err != nil {
                log.Printf("Failed to close stream: %v", err)
            }

			endTime := time.Now()
			duration := endTime.Sub(startTime)
			throughput := float64(totalBytes) / duration.Seconds()
			jitter := calculateJitter(delays)

			log.Printf("Total bytes: %d, Duration: %v, Throughput: %f Bps, Jitter: %v", totalBytes, duration, throughput, jitter)

        }()
    })

    log.Fatal(wt.ListenAndServeTLS("localhost.pem", "localhost-key.pem"))
}

func getFiles(dir string) ([]string, error) {
    var files []string
    for i := 0; i <= 30; i++ {
        file := filepath.Join(dir, fmt.Sprintf("frame_%d.jpg", i))
        files = append(files, file)
    }
    return files, nil
}

func sendFile(filename string, stream webtransport.Stream, fileSize *int) error {
    file, err := os.Open(filename)
	fileInfo, err := file.Stat()
	*fileSize = int(fileInfo.Size())
    if err != nil {
        return fmt.Errorf("failed to open file %s: %v", filename, err)
    }
    defer file.Close()

    _, err = io.Copy(stream, file)
    if err != nil {
        return fmt.Errorf("failed to send file %s: %v", filename, err)
    }
    return nil
}

func calculateJitter(delays []time.Duration) time.Duration {
	if len(delays) < 2 {
		return 0 // Jitter 계산 불가
	}
	var sum time.Duration
	var count int
	previousDelay := delays[0]
	for _, currentDelay := range delays[1:] {
		sum += time.Duration(math.Abs(float64(currentDelay - previousDelay)))
		previousDelay = currentDelay
		count++
	}
	return sum / time.Duration(count) // 평균 Jitter 반환
}
