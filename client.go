package torrent

import (
	"crypto/sha1"
	"strconv"
	"torrent/bencoding"	
)

type Client struct {
	peerId string
	port int
}

func NewClient() *Client {
	output := new(Client)
	return output
}

func (this *Client) PeerId() string {
	if this.peerId != "" {
		return this.peerId
	}
	this.peerId = GeneratePeerId()
	return this.peerId
}

func (this *Client) Port() int {
	if this.port == 0 {
		this.port = RandomPort()
	}
	return this.port
}

func infoHash(metaInfo *bencoding.Any) []byte {
	hasher := sha1.New()
	encodedMetaInfo, _ := bencoding.Encode(metaInfo.AsDictionary["info"])
	hasher.Write(encodedMetaInfo)
	return hasher.Sum(nil)
}

func (this *Client) NewTrackerQuery(torr *Torrent, event string) TrackerQuery {	
	output := make(TrackerQuery)
	
	output["info_hash"] = string(infoHash(torr.MetaInfo()))
	output["peer_id"] = this.PeerId()
	output["port"] = strconv.Itoa(this.Port())
	output["downloaded"] = strconv.Itoa(torr.DownloadedSize())
	output["uploaded"] = strconv.Itoa(torr.UploadedSize())
	output["left"] = strconv.Itoa(torr.LeftSize())
	output["compact"] = "1"
	output["numwant"] = "50"
	if event != "" {
		output["event"] = event
	}
	
	return output
}

