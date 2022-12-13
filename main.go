package main

import (
	"errors"
	"fmt"
	"github.com/go-ping/ping"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// hostPath host file path
var hostPath string = "/etc/hosts"

// ipRegexExpr ip regex expression
var ipRegexExpr string = "((25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))\\.){3}(25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))"

func main() {
	addresses := readHosts()
	if len(addresses) == 0 {
		log.Println("hosts file is empty")
		return
	}
	wg := &sync.WaitGroup{}
	for _, addr := range addresses {
		wg.Add(1)
		addr := addr
		go func() {
			defer wg.Done()
			err := pingTest(addr)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}
	wg.Wait()
	log.Println("all address test is done")
}

// readHosts readt hosts from host file
func readHosts() []string {
	f, err := os.Open(hostPath)
	if err != nil {
		log.Fatalln(err)
	}
	contentBs, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	content := string(contentBs)
	lines := strings.Split(content, "\n")
	ipRegex, err := regexp.Compile(ipRegexExpr)
	if err != nil {
		log.Fatalln(err)
	}
	addresses := make([]string, 0)
	for _, line := range lines {
		if strings.Contains(line, "#") {
			continue
		}
		ipBs := ipRegex.Find([]byte(line))
		if len(ipBs) > 0 {
			addresses = append(addresses, string(ipBs))
		}
	}
	return addresses
}

// pingTest use ping to test net connective
func pingTest(address string) error {
	pinger := ping.New(address)
	pinger.Count = 5
	pinger.Timeout = time.Second
	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}

	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	err := pinger.Run()
	if err != nil {
		return err
	}
	if pinger.Statistics().PacketLoss > 0.5 {
		// if packet loss rate bigger than 50%, then throw error
		return errors.New(fmt.Sprintf("address:%s can not conn", address))
	}
	return err
}
