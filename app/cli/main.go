package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/odit-bit/webcrawler"
	"github.com/odit-bit/webcrawler/x/xpipe"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	ui := userInput{scn: scanner}
	print := printer{}

	cr := webcrawler.NewCrawler(&ui, &print)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("input url ex: https://go.dev \n")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cr.Crawl(ctx)
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
