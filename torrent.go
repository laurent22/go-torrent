package torrent

import (
	"fmt"
	"errors"
	"torrent/bencoding"
)

type Torrent struct {
	url string
	metaInfo *bencoding.Any
	client *Client
}

func (this *Client) NewTorrent(url string) *Torrent {
	output := new(Torrent)
	output.url = url
	return output
}

func (this *Torrent) Url() string {
	return this.url
}

func (this *Torrent) DownloadedCount() int {
	return 0 // TODO
}

func (this *Torrent) UploadedCount() int {
	return 0 // TODO
}

func (this *Torrent) LeftCount() int {
	return 0 // TODO
}

func (this *Torrent) MetaInfo() *bencoding.Any {
	return this.metaInfo
}

func (this *Torrent) FetchMetaInfo() error {
	body, err := httpGet(this.Url(), NewHttpCallOptions())
	if err != nil { return err }
	metaInfo, err := bencoding.Decode(body)
	if err != nil { return err }
	this.metaInfo = metaInfo
	return nil
}

func (this *Torrent) CallTracker(query TrackerQuery) (*bencoding.Any, error) {
	announceUrl := this.MetaInfo().AsDictionary["announce"].AsString
	callUrl := httpGetUrl(announceUrl, map[string]string(query))
	body, err := httpGet(callUrl, NewHttpCallOptions())
	fmt.Println(callUrl)
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