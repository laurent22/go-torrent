package torrent

// TODO: seed random num generator

import (
	"fmt"
	"crypto/sha1"
	"errors"
	"math/rand"
	"strings"
	"torrent/bencoding"
)

const (
	ClientId = "GT"
	Version = "0.0.01"
)

var ErrInvalidBencodedData = errors.New("invalid bencoded data")

type TrackerQuery map[string]string

func peerIdPrefix() string {
	return "-" + ClientId + strings.Replace(Version, ".", "", -1) + "-"	
}

func peerIdSuffix() string {
	allowedChars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	output := ""
	for i := 0; i < 12; i++ {
		n := rand.Intn(len(allowedChars))
		output += string(allowedChars[n])
	}
	return output
}

func GeneratePeerId() string {
	return peerIdPrefix() + peerIdSuffix()
}

func RandomPort() int {
	return 10000 + rand.Intn(55000)
}