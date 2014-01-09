package torrent

import (
	"net/url"
	"strconv"	
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

func (this *Client) NewTrackerQuery(torr *Torrent) TrackerQuery {	
	output := make(TrackerQuery)
	
	output["info_hash"] = url.QueryEscape(string(metaInfoHash(torr.MetaInfo())))
	output["peer_id"] = url.QueryEscape(this.PeerId())
	output["port"] = strconv.Itoa(this.Port())
	output["downloaded"] = strconv.Itoa(torr.DownloadedCount())
	output["uploaded"] = strconv.Itoa(torr.UploadedCount())
	output["left"] = strconv.Itoa(torr.LeftCount())
	output["compact"] = "1"
	output["numwant"] = "50"
	
	return output
}

