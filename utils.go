package main

import (
	"sync"

	"github.com/iancmcc/goupnp"
)

func mergeDevices(cs ...<-chan goupnp.MaybeRootDevice) <-chan goupnp.MaybeRootDevice {
	var wg sync.WaitGroup
	out := make(chan goupnp.MaybeRootDevice)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan goupnp.MaybeRootDevice) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
