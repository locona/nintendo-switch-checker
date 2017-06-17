package nschecker

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Source struct {
	Name          string
	URL           string
	SoldOutText   string
	AvailableText string
}

type State int

const (
	UNKNOWN State = iota
	SOLDOUT
	AVAILABLE
	ERROR
)

func (s State) String() string {
	switch s {
	case UNKNOWN:
		return "UNKNOWN"
	case SOLDOUT:
		return "SOLDOUT"
	case AVAILABLE:
		return "AVAILABLE"
	case ERROR:
		return "ERROR"
	}
	return "Unknown state"
}

func Check(s Source) (State, error) {
	resp, err := http.Get(s.URL)
	if err != nil {
		return ERROR, err
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "charset=Windows-31J") {
		buf := new(bytes.Buffer)
		convertShiftJIS(resp.Body, buf)
		reader = buf
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if s.SoldOutText != "" && strings.Contains(text, s.SoldOutText) {
			return SOLDOUT, nil
		}
		if s.AvailableText != "" && strings.Contains(text, s.AvailableText) {
			return AVAILABLE, nil
		}
	}
	if s.AvailableText != "" {
		return SOLDOUT, nil
	}
	return AVAILABLE, nil
}

func convertShiftJIS(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, transform.NewReader(in, japanese.ShiftJIS.NewDecoder()))
	return err
}