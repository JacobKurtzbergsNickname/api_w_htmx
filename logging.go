package main

import (
	"io"
	"sync"

	"github.com/olekukonko/tablewriter"
)

// TableLogger manages the tablewriter and synchronization.
type TableLogger struct {
	writer *tablewriter.Table
	mutex  sync.Mutex
}

// NewTableLogger initializes the TableLogger with headers.
func NewTableLogger(output io.Writer) *TableLogger {
	table := tablewriter.NewWriter(output)
	table.SetHeader([]string{"Emoji", "Method", "Path", "Remote Addr"})
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)

	return &TableLogger{
		writer: table,
	}
}

// LogRequest appends a new row to the table in a thread-safe manner.
func (tl *TableLogger) LogRequest(emoji, method, path, remoteAddr string) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	tl.writer.Append([]string{emoji, method, path, remoteAddr})
	tl.writer.Render()
	tl.writer.ClearRows() // Clear rows after rendering to prepare for the next entry
}
