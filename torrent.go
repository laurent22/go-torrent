package torrent

import (
	"errors"
	"sort"
	"torrent/bencoding"
)

type Torrent struct {
	url string
	metaInfo *bencoding.Any
	client *Client
	selectedFileIndexes []int
	fileCount int
}

func (this *Client) NewTorrent(url string) *Torrent {
	output := new(Torrent)
	output.url = url
	return output
}

func (this *Torrent) FileCount() int {
	return this.fileCount
}

func (this *Torrent) SelectedFileIndexes() []int {
	return this.selectedFileIndexes
}

func (this *Torrent) SetSelectedFileIndexes(selection []int) error {
	sort.Ints(selection)
	previous := -1
	for i := 0; i < len(selection); i++ {
		index := selection[i]
		if index < 0 { return ErrIndexOutOfBound }
		if index >= this.fileCount { return ErrIndexOutOfBound }
		if previous == index { return ErrFileSelectionDuplicateIndex }
		previous = index
	}
	this.selectedFileIndexes = selection
	return nil
}

func (this *Torrent) Url() string {
	return this.url
}

func (this *Torrent) DownloadedSize() int {
	return 0 // TODO
}

func (this *Torrent) UploadedSize() int {
	return 0 // TODO
}

func (this *Torrent) LeftSize() int {
	return this.SelectedFileSize() - this.DownloadedSize()
}

func (this *Torrent) FileIndexIsSelected(index int) bool {
	for _, i := range this.selectedFileIndexes {
		if i == index { return true }
	}
	return false
}

func (this *Torrent) SelectedFileSize() int {
	info := this.MetaInfo().AsDictionary["info"].AsDictionary
	
	if this.IsSingleFile() {
		return info["length"].AsInt
	}
	
	output := 0
	for i, dic := range info["files"].AsList {
		if this.FileIndexIsSelected(i) {
			output += dic.AsDictionary["length"].AsInt
		}
	}
	return output	
}

func (this *Torrent) TotalFileSize() int {
	info := this.MetaInfo().AsDictionary["info"].AsDictionary
	
	if this.IsSingleFile() {
		return info["length"].AsInt
	}
	
	output := 0
	for _, dic := range info["files"].AsList {
		output += dic.AsDictionary["length"].AsInt
	}
	return output
}

func (this *Torrent) IsSingleFile() bool {
	info := this.MetaInfo().AsDictionary["info"].AsDictionary
	_, hasMultipleFiles := info["files"]
	return !hasMultipleFiles	
}

func (this *Torrent) MetaInfo() *bencoding.Any {
	return this.metaInfo
}

func (this *Torrent) initializeSelectedFileIndexes() {
	if this.IsSingleFile() {
		this.selectedFileIndexes = make([]int, 0, 1)
		this.selectedFileIndexes = append(this.selectedFileIndexes, 0)
	} else {
		files := this.MetaInfo().AsDictionary["info"].AsDictionary["files"].AsList
		this.selectedFileIndexes = make([]int, 0, len(files))
		for i, _ := range files {
			this.selectedFileIndexes = append(this.selectedFileIndexes, i)
		}
	}
	
	this.fileCount = len(this.selectedFileIndexes)
}

func (this *Torrent) FetchMetaInfo() error {
	body, err := httpGet(this.Url(), NewHttpCallOptions())
	if err != nil { return err }
	metaInfo, err := bencoding.Decode(body)
	if err != nil { return err }
	this.metaInfo = metaInfo
	this.initializeSelectedFileIndexes()
	return nil
}

func (this *Torrent) CallTracker(query TrackerQuery) (*bencoding.Any, error) {
	announceUrl := this.MetaInfo().AsDictionary["announce"].AsString
	callUrl := httpGetUrl(announceUrl, map[string]string(query))
	body, err := httpGet(callUrl, NewHttpCallOptions())
	if err != nil {
		return nil, err
	}
	output, err := bencoding.Decode(body)
	if err != nil {
		return output, err
	}
	// Check that the response is a bencoded dictionary and whether
	// it includes the "failure reason" key. If it does, it's an error.
	if output.Type != bencoding.Dictionary {
		return output, ErrInvalidBencodedData
	}
	failureReason, ok := output.AsDictionary["failure reason"]
	if ok {
		return output, errors.New(failureReason.AsString)
	}
	return output, nil
}

func (this *Torrent) TrackerUpdate() {
	
}