package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
)

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("no executable name is given")
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		prodCnt = flags.Uint("prods", 1, "the number of producers")
		consCnt = flags.Uint("conss", 1, "the number of consumers")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %s", err)
	}

	return runProdCons(w, flags.Args(), *prodCnt, *consCnt)
}

func runProdCons(w io.Writer, jobs []string, prodCnt, consCnt uint) error {
	prodJobs, consJobs := make(chan string), make(chan string)

	var prods sync.WaitGroup
	for i := uint(0); i < prodCnt; i++ {
		prods.Add(1)
		p := worker{
			WaitGroup: &prods,
			queue:     prodJobs,
			handle: func() func(string) {
				return func(j string) {
					consJobs <- j
				}
			}(),
		}

		go p.work()
	}

	var conss sync.WaitGroup
	for i := uint(0); i < consCnt; i++ {
		conss.Add(1)
		c := worker{
			WaitGroup: &conss,
			queue:     consJobs,
			handle: func(j string) {
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
