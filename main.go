package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	peer "github.com/libp2p/go-libp2p-peer"

	config "github.com/ipfs/go-ipfs/repo/config"
	ci "github.com/libp2p/go-libp2p-crypto"
)

var times = 0
var bitSize = 1024

func generatePeerID(toMatch string, c chan bool) {
	times = times + 1

	// TODO guard higher up
	ident := config.Identity{}

	sk, pk, err := ci.GenerateKeyPair(ci.RSA, bitSize)
	if err != nil {
		panic(err)
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := sk.Bytes()
	if err != nil {
		panic(err)
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		panic(err)
	}
	ident.PeerID = id.Pretty()

	matches := strings.Contains(strings.ToLower(ident.PeerID), strings.ToLower(toMatch))
	if matches {
		fmt.Println(ident.PeerID + " matched! ")
		spew.Dump(ident)
		c <- true
	} else {
		timeStr := strconv.Itoa(times)
		fmt.Println(timeStr + " " + ident.PeerID + " did not match...")
		c <- false
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
