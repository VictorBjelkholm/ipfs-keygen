package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	peer "github.com/libp2p/go-libp2p-peer"

	config "github.com/ipfs/go-ipfs/repo/config"
	ci "github.com/libp2p/go-libp2p-crypto"
)

var times = 0
var lastTimes = 0

var bitSize = 1024

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func saveIdentity(conf config.Identity) error {
	f, err := os.Create(path.Join("keys", conf.PeerID))
	if err != nil {
		return err
	}
	e := json.NewEncoder(f)
	err = e.Encode(conf)
	if err != nil {
		return err
	}
	return f.Close()
}

func try(ctx context.Context, toMatch string, n int, successes chan<- config.Identity) {
	for i := 0; i < n; i++ {
		fmt.Println("Starting")
		go func() {
			for {
				conf, err := generatePeerID(toMatch)
				if err != nil {
					continue
				}
				matches := strings.Contains(strings.ToLower(conf.PeerID), strings.ToLower(toMatch))
				if matches {
					successes <- conf
				}

				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}
}

func generatePeerID(toMatch string) (config.Identity, error) {
	times = times + 1

	// TODO guard higher up
	ident := config.Identity{}

	sk, pk, err := ci.GenerateKeyPair(ci.RSA, bitSize)
	// sk, pk, err := ci.GenerateKeyPair(ci.Secp256k1, bitSize)
	if err != nil {
		return ident, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := sk.Bytes()
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()
	return ident, err

	// matches := strings.Contains(strings.ToLower(ident.PeerID), strings.ToLower(toMatch))
	// if matches {
	// 	fmt.Println(ident.PeerID + " matched and saved! ")
	// 	saveIdentity(ident)
	// 	return false, true
	// 	panic(err)
	// } else {
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return false, false
	// }
}

func start(toMatch string, workers int) {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			timeStr := strconv.Itoa(times)
			timesSinceLastSecond := strconv.Itoa(times - lastTimes)
			lastTimes = times
			fmt.Println("Done " + timeStr + " tries so far, " + timesSinceLastSecond + " per second")
		}
	}()
	fmt.Println("## Searching for '" + toMatch + "'")

	ctx, cancel := context.WithCancel(context.Background())

	successes := make(chan config.Identity)
	fmt.Println(workers)
	go try(ctx, toMatch, workers, successes)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case <-c:
			cancel()
			return
		case conf := <-successes:
			if err := saveIdentity(conf); err != nil {
				panic(err)
			}
			fmt.Printf("saved %s\n", conf.PeerID)
		}
	}
}

var numWorkers int
var maxProcs int
var toMatch string

func main() {
	flag.IntVar(&numWorkers, "workers", runtime.NumCPU(), "Number of workers")
	flag.IntVar(&maxProcs, "procs", runtime.NumCPU(), "Number of GO MAX PROCS")
	flag.StringVar(&toMatch, "match", "ipfs", "What to match against")
	flag.Parse()
	runtime.GOMAXPROCS(maxProcs)
	start(toMatch, numWorkers)

	// matched := false
	// for !matched {
	// 	matched = generatePeerID(toMatch)
	// }
}
