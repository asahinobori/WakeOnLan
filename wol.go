package main

import (
	"errors"
	"flag"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

var (
	log        *logrus.Logger
	mPacket    *magicPacket
	targetHost string
	targetMAC  string
	targetPort int
)

type magicPacket [102]byte

func init() {
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	})

	// fix output format for windows os
	if runtime.GOOS == "windows" {
		log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	}

	targetHost = "your.domain.org"
	targetMAC = "AA:BB:CC:DD:EE:FF"
	targetPort = 9
}

func newMagicPacket(macAddr string) (mp *magicPacket, err error) {
	mac, err := net.ParseMAC(macAddr)
	if err != nil {
		return nil, err
	}

	if len(mac) != 6 {
		return nil, errors.New("invalid EUI-48 MAC address")
	}

	mPacket := &magicPacket{}

	copy(mPacket[0:], []byte{255, 255, 255, 255, 255, 255})
	offset := 6

	for i := 0; i < 16; i++ {
		copy(mPacket[offset:], mac)
		offset += 6
	}

	return mPacket, nil
}

func sendUDPPacket(mPacket *magicPacket, addr string) (err error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(mPacket[:])
	return err
}

func main() {
	flag.StringVar(&targetHost, "host", targetHost, "host domain or broadcast address")
	flag.StringVar(&targetMAC, "mac", targetMAC, "mac address, ex:BB:CC:DD:EE:FF:11")
	flag.IntVar(&targetPort, "port", targetPort, "host port")
	flag.Parse()

	log.WithFields(logrus.Fields{
		"targetHost": targetHost,
		"targetMAC":  targetMAC,
		"targetPort": targetPort,
	}).Info("params")

	targetIps, err := net.LookupHost(targetHost)
	if err != nil {
		log.WithField("error", err).WithField("targetHost", targetHost).Error("lookup target host failed")
		return
	}

	if mPacket, err = newMagicPacket(targetMAC); err != nil {
		log.WithField("error", err).WithField("targetMAC", targetMAC).Error("new magic packet failed")
		return
	}

	if len(targetIps) != 0 && mPacket != nil {
		err = sendUDPPacket(mPacket, targetIps[0]+":"+strconv.Itoa(targetPort))
		if err != nil {
			log.WithField("error", err).Error("send magic packet failed")
			return
		}
	} else {
		log.WithField("targetHost", targetHost).Error("lookup target host return empty ip list")
		return
	}

	log.WithField("targetIp", targetIps[0]).Info("send magic packet sucessfully")
}
