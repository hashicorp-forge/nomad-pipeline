package coordinator

import (
	"context"

	"github.com/hpcloud/tail"
)

type LogStream struct {
	path     string
	errorCh  chan error
	streamCh chan string
}

func NewLogStream(path string) *LogStream {
	return &LogStream{
		path:     path,
		errorCh:  make(chan error),
		streamCh: make(chan string),
	}
}

func (s *LogStream) ErrorCh() <-chan error { return s.errorCh }

func (s *LogStream) StreamCh() <-chan string { return s.streamCh }

func (s *LogStream) Run(ctx context.Context) {

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
