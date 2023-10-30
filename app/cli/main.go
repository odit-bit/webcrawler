package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/odit-bit/webcrawler"
	"github.com/odit-bit/webcrawler/x/xpipe"
)

func main() {

	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	// Start CPU profiling

	var pprofStr string
	flag.StringVar(&pprofStr, "pprof", "", "profiling ")
	flag.Parse()

	fmt.Println(pprofStr)
	if pprofStr != "" {
		switch pprofStr {
		case "cpu":
			f, _ := os.Create("cpu.pprof")
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()

		case "memory":
			f, _ := os.Create("memory.pprof")
			defer f.Close() // error handling omitted for example
			runtime.GC()    // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
		default:
			fmt.Println("no profiling")
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	ui := userInput{scn: scanner}
	print := printer{}

	cr := webcrawler.NewCrawler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("input url ex: https://go.dev \n")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cr.Crawl(ctx, &ui, &print)
		if err != nil {
			log.Println("crawler error:", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()
	fmt.Println(ui.Error())
	log.Println("exit app")
}

var _ xpipe.Fetcher[*webcrawler.Resource] = (*userInput)(nil)

type userInput struct {
	scn *bufio.Scanner
}

// Error implements xpipe.Fetcher.
func (ui *userInput) Error() error {
	return ui.scn.Err()
}

// Next implements xpipe.Fetcher.
func (ui *userInput) Next() bool {
	return ui.scn.Scan()
}

// Resource implements xpipe.Fetcher.
func (ui *userInput) Resource() *webcrawler.Resource {
	text := ui.scn.Text()
	wr := webcrawler.Resource{
		URL: text,
	}

	return &wr
}

var _ xpipe.Streamer[*webcrawler.Resource] = (*printer)(nil)

type printer struct {
}

// Consume implements xpipe.Streamer.
func (p *printer) Consume(ctx context.Context, result <-chan *webcrawler.Resource) error {
	for {
		select {
		case <-ctx.Done():
			//break
		case r, ok := <-result:
			if ok {
				fmt.Println(r)
				continue
			}
			// break
		}
		break
	}
	return nil
}
