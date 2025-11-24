package coordinator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func (c *Coordinator) Getlogs(namespace, runID, stepID, logType string) ([]string, error) {

	var lines []string

	fileHandle, err := os.Open(logPath(c.dataDir, namespace, runID, stepID, logType))
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

func (c *Coordinator) StreamLogs(namespace, runID, stepID, logType string) *LogStream {
	return NewLogStream(logPath(c.dataDir, namespace, runID, stepID, logType))
}

func (c *Coordinator) WriteLogsBatch(namespace, runID, stepID, logType string, lines []string) error {

	path := logPath(c.dataDir, namespace, runID, stepID, logType)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	totalBytes := 0

	// Write the log bytes
	for _, line := range lines {
		bytesWritten, err := f.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("failed to write log batch: %w", err)
		}
		totalBytes += bytesWritten
	}

	c.logger.Debug("successfully wrote log batch to disk",
		zap.String("run_id", runID),
		zap.String("step_id", stepID),
		zap.String("type", logType),
		zap.Int("bytes", totalBytes))

	return nil
}

func logPath(dataDir, namespace, runID, stepID, logType string) string {
	return filepath.Join(logDir(dataDir, namespace, runID, stepID), "logs", fmt.Sprintf("%s.log", logType))
}

func logDir(dataDir, namespace, runID, stepID string) string {
	return filepath.Join(dataDir, namespace, runID, stepID)
}
