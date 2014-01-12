package torrent

import (
	"torrent/bencoding"
	"net/http"
	"testing"
)

var sampleTorrentTrackerUrl = "http://localhost:8080/LibreOffice.torrent"
var sampleTorrentAnnounceUrl = "http://tracker.documentfoundation.org:6969/announce"

func startTestServer() {
	http.ListenAndServe(":8080", http.FileServer(http.Dir("testing")))
}

func Test_FetchMetaInfo(t *testing.T) {
	go startTestServer()
	
	client := NewClient()
	torr := client.NewTorrent(sampleTorrentTrackerUrl)	
	err := torr.FetchMetaInfo()
	
	if err != nil {
		t.Fatal("Error fetching meta info:", err)
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
	
	hash := infoHash(metaInfo)
	if len(hash) != 20 {
		t.Errorf("Expected a length of %d, got %d", 20, len(hash))
	}
}

func Test_NewTrackerQuery(t *testing.T) {
	go startTestServer()
	
	client := NewClient()
	torr := client.NewTorrent(sampleTorrentTrackerUrl)	
	torr.FetchMetaInfo()
	
	query := client.NewTrackerQuery(torr, "")
	
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

func Test_HttpGetUrl(t *testing.T) {
	type GetUrlTest struct {
		baseUrl string
		parameters map[string]string
		output string
	}
	
	var tests = []GetUrlTest{
		{ "http://test.com", map[string]string{"one":"123"}, "http://test.com?one=123" },
		{ "http://test.com", map[string]string{}, "http://test.com" },
		{ "http://test.com", map[string]string{"one":"123","two":"abcd","first":"555"}, "http://test.com?first=555&one=123&two=abcd" },
		{ "http://test.com", map[string]string{"enc":"ab cdé"}, "http://test.com?enc=ab+cd%C3%A9" },
	}
	
	for _, d := range tests {
		output := httpGetUrl(d.baseUrl, d.parameters)
		if output != d.output { t.Errorf("Expected \"%s\", got \"%s\"", d.output, output) }
	}
}

func Test_TotalFileSize(t *testing.T) {
	go startTestServer()
	
	type TotalFileSizeTest struct {
		url string
		expected int
	}
	 
	var tests = []TotalFileSizeTest{
		{ "http://localhost:8080/Despicable Me (2010) [1080p].torrent", 1549245214 },
		{ "http://localhost:8080/The Cure - Disintegration [1989] (320 Kbps) [Dodecahedron].torrent", 177220781 },
		{ "http://localhost:8080/The Pretenders - Break Up The Concrete (Advance) [2008] - Rock [www.torrentazos.com].torrent", 50630290 },
		{ "http://localhost:8080/LibreOffice.torrent", 181549113 },
	}
	
	client := NewClient()
	
	for _, d := range tests {
		torr := client.NewTorrent(d.url)
		torr.FetchMetaInfo()
		output := torr.TotalFileSize()
		if output != d.expected { t.Errorf("Expected \"%s\", got \"%s\"", d.expected, output) }
	}
}

func Test_IsSingleFile(t *testing.T) {
	go startTestServer()
	
	type IsSingleFileTest struct {
		url string
		expected bool
	}
	 
	var tests = []IsSingleFileTest{
		{ "http://localhost:8080/Despicable Me (2010) [1080p].torrent", false },
		{ "http://localhost:8080/The Cure - Disintegration [1989] (320 Kbps) [Dodecahedron].torrent", false },
		{ "http://localhost:8080/LibreOffice.torrent", true },
	}
	
	client := NewClient()
	
	for _, d := range tests {
		torr := client.NewTorrent(d.url)
		torr.FetchMetaInfo()
		if torr.IsSingleFile() != d.expected { t.Errorf("Expected \"%s\", got \"%s\"", d.expected, torr.IsSingleFile()) }
	}	
}

func Test_SelectedFileIndexes(t *testing.T) {
	go startTestServer()
	
	type SelectedFileIndexesTest struct {
		url string
		initial int
	}
	 
	var tests = []SelectedFileIndexesTest{
		{ "http://localhost:8080/The Pretenders - Break Up The Concrete (Advance) [2008] - Rock [www.torrentazos.com].torrent", 15 },
		{ "http://localhost:8080/LibreOffice.torrent", 1 },
	}
	
	client := NewClient()
	var err error
	
	for _, d := range tests {
		torr := client.NewTorrent(d.url)
		torr.FetchMetaInfo()
		indexes := torr.SelectedFileIndexes()
		if len(indexes) != d.initial { t.Errorf("Expected \"%s\", got \"%s\"", d.initial, len(indexes)) }
		if torr.FileCount() != d.initial { t.Errorf("Expected \"%s\", got \"%s\"", d.initial, torr.FileCount()) }
		
		for i, index := range indexes {
			if i != index { t.Errorf("Expected \"%s\", got \"%s\"", i, index) }			
		}
		
		err = torr.SetSelectedFileIndexes([]int{0,torr.FileCount() + 1})
		if err != ErrIndexOutOfBound { t.Errorf("Expected \"%s\", got \"%s\"", ErrIndexOutOfBound, err) }

		err = torr.SetSelectedFileIndexes([]int{0})
		if err != nil { t.Errorf("Expected no error, got \"%s\"", err) }
		
		err = torr.SetSelectedFileIndexes([]int{0,0,0})
		if err != ErrFileSelectionDuplicateIndex { t.Errorf("Expected \"%s\", got \"%s\"", ErrFileSelectionDuplicateIndex, err) }
	}
}