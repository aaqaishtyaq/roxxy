package log

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"gopkg.in/check.v1"
)

type LogSuite struct{}

var _ = check.Suite(&LogSuite{})

func Test(t *testing.T) {
	check.TestingT(t)
}

type nopCloseWriter struct{ io.Writer }

func (nopCloseWriter) Close() error { return nil }

func (s *LogSuite) TestNewFileLogger(c *check.C) {
	file, err := ioutil.TempFile("", "loggettest")
	c.Assert(err, check.IsNil)
	file.Close()
	fileName := file.Name()
	defer os.Remove(fileName)
	logger, err := NewFileLogger(fileName)
	c.Assert(err, check.IsNil)
	_, err = os.Stat(fileName)
	c.Assert(err, check.IsNil)
	logger.MessageRaw(&LogEntry{})
	logger.Stop()
	data, err := ioutil.ReadFile(fileName)
	c.Assert(err, check.IsNil)
	c.Assert(string(data), check.Equals, "::ffff: - - [Mon Jan  1 00:00:00 UTC 0001] \"  \" 0 0 \"\" \"\" \":\" \"\" \"\" 0.000 0.000\n")
}

func (s *LogSuite) TestNewSyslogLogger(c *check.C) {
	logger, err := NewSyslogLogger()
	c.Assert(err, check.IsNil)
	logger.MessageRaw(&LogEntry{})
	logger.Stop()
}

func (s *LogSuite) TestNewStdoutLogger(c *check.C) {
	logger, err := NewStdoutLogger()
	c.Assert(err, check.IsNil)
	logger.MessageRaw(&LogEntry{})
	logger.Stop()
}

func (s *LogSuite) TestNewWriterLogger(c *check.C) {
	buffer := &bytes.Buffer{}
	logger := NewWriterLogger(nopCloseWriter{buffer})
	logger.MessageRaw(&LogEntry{})
	logger.Stop()
	c.Assert(buffer.String(), check.Equals, "::ffff: - - [Mon Jan  1 00:00:00 UTC 0001] \"  \" 0 0 \"\" \"\" \":\" \"\" \"\" 0.000 0.000\n")
}

func (s *LogSuite) TestLoggerMessageAfterStop(c *check.C) {
	buffer := &bytes.Buffer{}
	logger := NewWriterLogger(nopCloseWriter{buffer})
	logger.MessageRaw(&LogEntry{})
	logger.Stop()
	logger.MessageRaw(&LogEntry{})
	c.Assert(buffer.String(), check.Equals, "::ffff: - - [Mon Jan  1 00:00:00 UTC 0001] \"  \" 0 0 \"\" \"\" \":\" \"\" \"\" 0.000 0.000\n")
}

func (s *LogSuite) TestLoggerFull(c *check.C) {
	buffer := &bytes.Buffer{}
	ch := make(chan time.Time)
	close(ch)
	logger := Logger{
		logChannel: make(chan *LogEntry, 1),
		done:       make(chan struct{}),
		wg:         &sync.WaitGroup{},
		writer:     nopCloseWriter{buffer},
		nextNotify: ch,
	}
	logger.wg.Add(1)
	logger.MessageRaw(&LogEntry{})
	logger.MessageRaw(&LogEntry{})
	logger.MessageRaw(&LogEntry{})
	logger.MessageRaw(&LogEntry{})
	go logger.logWriter()
	logger.Stop()
	c.Assert(buffer.String(), check.Equals, `::ffff: - - [Mon Jan  1 00:00:00 UTC 0001] "  " 0 0 "" "" ":" "" "" 0.000 0.000
Dropping log messages to due to full channel buffer.
`)
}
