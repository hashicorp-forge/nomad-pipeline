package logs

import (
	"bufio"
	"context"
	"os"

	"github.com/hpcloud/tail"
)

type Stream struct {
	path     string
	errorCh  chan<- error
	streamCh chan<- string
}

func NewStream(path string, streamCh chan<- string, errorCh chan<- error) *Stream {
	return &Stream{
		path:     path,
		errorCh:  errorCh,
		streamCh: streamCh,
	}
}

func (s *Stream) Run(ctx context.Context) {

	fileTail, err := tail.TailFile(s.path,
		tail.Config{
			Follow: true,
			Location: &tail.SeekInfo{
				Offset: 0,
				Whence: 0,
			},
			Logger: tail.DiscardingLogger,
		},
	)
	if err != nil {
		s.errorCh <- err
		return
	}

	defer func(fileTail *tail.Tail) { _ = fileTail.Stop() }(fileTail)

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-fileTail.Lines:
			if line != nil {
				s.streamCh <- line.Text
			}
		}
	}
}

func Get(file string) ([]string, error) {

	var lines []string

	fileHandle, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
