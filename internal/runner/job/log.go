package job

import (
	"bufio"
	"context"
	"io"
	"net/rpc"
	"time"

	"go.uber.org/zap"

	sharedrpc "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/rpc"
)

const (
	logBufferLimit   = 50
	logIntervalLimit = 5 * time.Second
)

type LogHandlerReq struct {
	Namespace string
	RunID     string
	StepID    string
	Type      string
}

type LogHandler struct {
	req       *LogHandlerReq
	logger    *zap.Logger
	rpcClient *rpc.Client

	buffer []string

	cmdPipe io.ReadCloser
}

func NewLogHandler(logger *zap.Logger, pipe io.ReadCloser, rpcClient *rpc.Client, req *LogHandlerReq) *LogHandler {
	return &LogHandler{
		req:       req,
		logger:    logger.Named("logs").With(zap.String("type", req.Type)),
		buffer:    []string{},
		cmdPipe:   pipe,
		rpcClient: rpcClient,
	}
}

func (l *LogHandler) Start(ctx context.Context) {
	l.logger.Info("starting log handler")
	l.reader(ctx)
	l.logger.Info("stopping log handler")
}

func (l *LogHandler) reader(ctx context.Context) {

	defer l.flushLogs()

	buf := bufio.NewScanner(l.cmdPipe)

	ticker := time.NewTicker(logIntervalLimit)
	defer ticker.Stop()

	for buf.Scan() {

		select {
		case <-ctx.Done():
			return
		default:
		}

		line := buf.Text()
		l.buffer = append(l.buffer, line)

		select {
		case <-ticker.C:
			l.flushLogs()
		default:
			if len(l.buffer) >= logBufferLimit {
				l.flushLogs()
			}
		}
	}
}

func (l *LogHandler) flushLogs() {

	if len(l.buffer) == 0 {
		return
	}

	logLines := make([]string, len(l.buffer))
	copy(logLines, l.buffer)

	l.buffer = l.buffer[:0]

	req := sharedrpc.RunnerLogsBatchReq{
		Namespace: l.req.Namespace,
		RunID:     l.req.RunID,
		StepID:    l.req.StepID,
		Type:      l.req.Type,
		Logs:      logLines,
	}

	if err := l.rpcClient.Call(sharedrpc.RunnerLogsBatchMethodName, req, nil); err != nil {
		l.logger.Error("failed to send log batch", zap.Error(err))
	} else {
		l.logger.Debug("successfully sent log batch", zap.Int("num_lines", len(logLines)))
	}
}
