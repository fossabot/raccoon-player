package main

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"time"
)

type cmdConfig struct {
	file     string
	url      string
	sleep    time.Duration
	pcap     bool
	udp      bool
	workers  int
	duration time.Duration
}

var (
	Cmd = &cobra.Command{
		Use:   "player",
		Short: "log player",
		Args:  cobra.ExactArgs(0),
		RunE:  run,
	}
	cfg     cmdConfig
	replays [][]byte
)

func init() {
	// Replay file path
	Cmd.Flags().StringVar(
		&cfg.file,
		"file",
		"",
		"replay file path")

	// Destination URL
	Cmd.Flags().StringVar(
		&cfg.url,
		"url",
		"127.0.0.1:1514",
		"destination URL")

	// Interval between message sends
	Cmd.Flags().DurationVar(
		&cfg.sleep,
		"sleep",
		0,
		"sleep duration between messages")

	// Replay file contains pcap dump
	Cmd.Flags().BoolVar(
		&cfg.pcap,
		"pcap",
		false,
		"whether replay file contains pcap dump")

	// Use UDP to connect to destination
	Cmd.Flags().BoolVar(
		&cfg.udp,
		"udp",
		false,
		"UDP mode")

	// Workers
	Cmd.Flags().IntVar(
		&cfg.workers,
		"workers",
		runtime.NumCPU(),
		"workers count")

	// Replay duration
	Cmd.Flags().DurationVar(
		&cfg.duration,
		"duration",
		time.Hour,
		"replay duration")

	_ = Cmd.MarkFlagRequired("file")
}

func main() {
	if err := Cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(_ *cobra.Command, _ []string) error {
	if !cfg.pcap {
		getReplaysFromTextFile()
	} else {
		getReplaysFromPcapFile()
	}

	fmt.Printf("* workers: %d\n* sleep: %s\n* destination: %s\n", cfg.workers, cfg.sleep, cfg.url)

	for i := 0; i < cfg.workers; i++ {
		go worker()
	}

	time.Sleep(cfg.duration)
	return nil
}

func worker() {
	proto := "tcp"
	if cfg.udp {
		proto = "udp"
	}

	connection, err := net.Dial(proto, cfg.url)
	if err != nil {
		log.Fatal(err)
	}

	for {
		for i := range replays {
			_, err := connection.Write(replays[i])
			if err != nil {
				log.Fatal(err)
			}
			if cfg.sleep > 0 {
				time.Sleep(cfg.sleep)
			}
		}
	}
}

func getReplaysFromTextFile() {
	data, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		log.Fatal(err)
	}

	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		replays = append(replays, line)
		replays = append(replays, []byte("\n"))
	}
}

func getReplaysFromPcapFile() {
	handle, err := pcap.OpenOffline(cfg.file)
	if err != nil {
		log.Fatal(err)
	}

	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		pld := packet.TransportLayer().LayerPayload()
		replays = append(replays, pld)
	}
}
