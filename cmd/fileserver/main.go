package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/judwhite/go-svc"
)

type program struct {
	wg           sync.WaitGroup
	quit         chan struct{}
	isWinService bool
}

func main() {
	prg := &program{}

	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	log.Printf("is win service? %v\n", env.IsWindowsService())

	p.isWinService = env.IsWindowsService()

	return nil
}

func (p *program) Start() error {
	p.quit = make(chan struct{})

	p.wg.Add(1)

	go func() {
		log.Println("Starting...")

		app := New(p.isWinService)
		if err := app.Run(); err != nil {
			log.Fatal(err)
			return
		}

		<-p.quit
		log.Println("Quit signal received...")

		if err := app.Close(); err != nil {
			fmt.Fprintf(os.Stdout, "Cannot graceful shutdown: %v\n", err)
		}

		p.wg.Done()
	}()

	return nil
}

func (p *program) Stop() error {
	log.Println("Stopping...")
	close(p.quit)
	p.wg.Wait()

	log.Println("Stopped.")

	return nil
}
