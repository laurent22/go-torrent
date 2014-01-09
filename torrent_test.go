package torrent

import (
	"../bencoding"
	"net/http"
	"testing"
)

var sampleTorrentTrackerUrl = "http://localhost:8080/LibreOffice.torrent"
var sampleTorrentAnnounceUrl = "http://tracker.documentfoundation.org:6969/announce"

func startTestServer() {
	http.ListenAndServe(":8080", http.FileServer(http.Dir("../testing")))
}

func Test_FetchMetaInfo(t *testing.T) {
	go startTestServer()
	
	client := NewClient()
	torr := client.NewTorrent(sampleTorrentTrackerUrl)	
	err := torr.FetchMetaInfo()
	
	if err != nil {
		t.Error("Error fetching meta info:", err)
	}
	
	metaInfo := torr.MetaInfo()
	if metaInfo.Type != bencoding.Dictionary {
		t.Error("Incorrect type:", metaInfo.Type)
	}
	
	metaInfoAnnounce, ok := metaInfo.AsDictionary["announce"]
	if !ok {
		t.Error("Doesn't have announce key")
	}
	
	if metaInfoAnnounce.AsString != sampleTorrentAnnounceUrl {
		t.Error("Invalid announce key")
	}
	
	hash := metaInfoHash(metaInfo)
	if len(hash) != 20 {
		t.Errorf("Expected a length of %d, got %d", 20, len(hash))
	}
}

func Test_NewTrackerQuery(t *testing.T) {
	go startTestServer()
	
	client := NewClient()
	torr := client.NewTorrent(sampleTorrentTrackerUrl)	
	torr.FetchMetaInfo()
	
	query := client.NewTrackerQuery(torr)
	
	requiredFields := []string{"info_hash","peer_id","port","downloaded","uploaded","left","compact","numwant"}
	for _, field := range requiredFields {
		_, ok := query[field]
		if !ok {
			t.Errorf("Required field '%s' is missing", field)
		}
	}
}

func Test_GeneratePeerId(t *testing.T) {
	var previous string
	for i := 0; i < 100; i++ {
		p := GeneratePeerId()
		if len(p) != 20 {
			t.Errorf("Length is %d instead of %d: '%s'", len(p), 20, p)
		}
		if previous == p {
			t.Errorf("Each peer ID should be unique: '%s' / '%s'", p, previous)			
		}
		previous = p
	}
	
	c := NewClient()
	p1 := c.PeerId()
	p2 := c.PeerId()
	if p1 != p2 {
		t.Errorf("The peer ID should not change within the same session: '%s' / '%s'", p1, p2)
	}
}

func Test_ClientPort(t *testing.T) {
	c := NewClient()
	p1 := c.Port()
	p2 := c.Port()
	if p1 <= 0 {
		t.Errorf("Expected port greated than zero, got %d", p1)
	}
	if p1 != p2 {
		t.Errorf("Successive calls to Port() should return the same port: %d / %d", p1, p2)		
	}
}