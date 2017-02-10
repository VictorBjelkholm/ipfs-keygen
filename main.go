package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	config "github.com/ipfs/go-ipfs/repo/config"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
)

var times = 0
var bitSize = 1024

func generatePeerID(toMatch string, c chan bool) {
	times = times + 1
	now := time.Now()
	directory := path.Join("/dev/shm", now.String())
	if !fsrepo.IsInitialized(directory) {
		cfg, err := config.Init(ioutil.Discard, bitSize)
		if err != nil {
			panic(err)
		}
		err = fsrepo.Init(directory, cfg)
		if err != nil {
			panic(err)
		}
	}
	r, err := fsrepo.Open(directory)
	defer r.Close()
	conf, err := r.Config()
	if err != nil {
		panic(err)
	}
	matches := strings.Contains(strings.ToLower(conf.Identity.PeerID), strings.ToLower(toMatch))
	if matches {
		fmt.Println(conf.Identity.PeerID + " matched! " + directory)
		c <- true
	} else {
		timeStr := strconv.Itoa(times)
		fmt.Println(timeStr + " " + conf.Identity.PeerID + " did not match...")
		c <- false
		err = os.RemoveAll(directory)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Need to pass search-term as first argument")
	}
	toMatch := os.Args[1]
	fmt.Println("## Searching for '" + toMatch + "'")
	matched := false
	ch1 := make(chan bool)
	for !matched {
		go generatePeerID(toMatch, ch1)
		matched = <-ch1
	}
}
