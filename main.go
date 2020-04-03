package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	var cnf config
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.UintVar(&cnf.prodCnt, "prods", 1, "the number of producers")
	flags.UintVar(&cnf.consCnt, "conss", 1, "the number of consumers")
	flags.DurationVar(&cnf.prodDelay, "prod-delay", 0, "fixed delay till a producer work")
	flags.DurationVar(&cnf.consDelay, "cons-delay", 0, "fixed delay till a consumer work")
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %s", err)
	}

	return runProdCons(w, flags.Args(), cnf)
}

func runProdCons(w io.Writer, jobs []string, cnf config) error {
	prodJobs, consJobs := make(chan string), make(chan string)

	var prods sync.WaitGroup
	for i := uint(0); i < cnf.prodCnt; i++ {
		prods.Add(1)
		p := worker{
			WaitGroup: &prods,
			queue:     prodJobs,
			handle: func() func(string) {
				return func(j string) {
					time.Sleep(cnf.prodDelay)
					consJobs <- j
				}
			}(),
		}

		go p.work()
	}

	var conss sync.WaitGroup
	for i := uint(0); i < cnf.consCnt; i++ {
		conss.Add(1)
		c := worker{
			WaitGroup: &conss,
			queue:     consJobs,
			handle: func(j string) {
				time.Sleep(cnf.consDelay)
				fmt.Fprintln(w, j)
			},
		}

		go c.work()
	}

	for _, j := range jobs {
		prodJobs <- j
	}
	close(prodJobs)

	prods.Wait()
	close(consJobs)

	conss.Wait()

	return nil
}

type config struct {
	prodCnt, consCnt     uint
	prodDelay, consDelay time.Duration
}

type worker struct {
	*sync.WaitGroup
	queue  <-chan string
	handle func(string)
}

func (w worker) work() {
	defer w.Done()

	for j := range w.queue {
		w.handle(j)
	}
}
