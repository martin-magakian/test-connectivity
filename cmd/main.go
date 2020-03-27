package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/kylegrantlucas/speedtest"
	"github.com/sparrc/go-ping"
)

// Generate Data Augmentor requests based on JSON logs
func main() {

	// read CLI
	pingScheduleSecond := flag.Int("ping", 5, "How ofter ping should run")
	speedtestScheduleSecond := flag.Int("speedtest", 120, "How ofter a speedtest should run")
	flag.Parse()

	// Handle shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// create logger
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	runPing(logger)
	runSpeedtest(logger)

	pingTicker := time.NewTicker(time.Second * time.Duration(*pingScheduleSecond))
	speedtestTicker := time.NewTicker(time.Second * time.Duration(*speedtestScheduleSecond))
	for {
		select {
		case <-pingTicker.C:
			runPing(logger)
		case <-speedtestTicker.C:
			runSpeedtest(logger)
		case <-stop:
			fmt.Println("Received stop signal stopping...")
			return
		}
	}
}

func runPing(logger *log.Logger) {
	pinger, err := ping.NewPinger("www.google.com")
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	if err != nil {
		panic(err)
	}
	pinger.Count = 3
	pinger.Run()                 // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats
	logger.Printf("[PING] sent %d ping in an avg of %d ms", stats.PacketsSent, stats.AvgRtt.Milliseconds())
}

func runSpeedtest(logger *log.Logger) {
	client, err := speedtest.NewDefaultClient()
	if err != nil {
		logger.Printf("[SPEEDTEST] error creating client: %v", err)
	}

	// Pass an empty string to select the fastest server
	server, err := client.GetServer("")
	if err != nil {
		logger.Printf("[SPEEDTEST] error getting server: %v", err)
	}

	dmbps, err := client.Download(server)
	if err != nil {
		logger.Printf("[SPEEDTEST] error getting download: %v", err)
	}

	umbps, err := client.Upload(server)
	if err != nil {
		logger.Printf("[SPEEDTEST] error getting upload: %v", err)
	}

	logger.Printf("[SPEEDTEST] Ping: %3.2f ms | Download: %3.2f Mbps | Upload: %3.2f Mbps\n", server.Latency, dmbps, umbps)
}
