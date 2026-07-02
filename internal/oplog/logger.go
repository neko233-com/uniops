package oplog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Entry represents a single operation log entry
type Entry struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	UserID    uint   `json:"user_id"`
	Action    string `json:"action"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	IP        string `json:"ip"`
	Detail    string `json:"detail,omitempty"`
}

// Logger writes operation logs to daily JSONL files
type Logger struct {
	dir string
	mu  sync.Mutex
}

// New creates a logger that writes to dir/YYYY-MM-DD.jsonl
func New(dir string) *Logger {
	os.MkdirAll(dir, 0755)
	return &Logger{dir: dir}
}

// Log writes an entry to today's log file
func (l *Logger) Log(entry Entry) {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	filename := filepath.Join(l.dir, time.Now().Format("2006-01-02")+".jsonl")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	data, _ := json.Marshal(entry)
	f.Write(data)
	f.Write([]byte("\n"))
}

// ListDates returns available log dates (newest first)
func (l *Logger) ListDates() ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, err
	}

	var dates []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			date := strings.TrimSuffix(e.Name(), ".jsonl")
			dates = append(dates, date)
		}
	}

	// Reverse for newest first
	for i, j := 0, len(dates)-1; i < j; i, j = i+1, j-1 {
		dates[i], dates[j] = dates[j], dates[i]
	}
	return dates, nil
}

// ReadDate reads all entries for a given date
func (l *Logger) ReadDate(date string) ([]Entry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	filename := filepath.Join(l.dir, date+".jsonl")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// DeleteRange deletes log files in [startDate, endDate] inclusive
func (l *Logger) DeleteRange(startDate, endDate string) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0, fmt.Errorf("invalid end date: %w", err)
	}

	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		dateStr := strings.TrimSuffix(e.Name(), ".jsonl")
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if !date.Before(start) && !date.After(end) {
			if err := os.Remove(filepath.Join(l.dir, e.Name())); err == nil {
				deleted++
			}
		}
	}
	return deleted, nil
}

// Search searches logs across all dates matching criteria
func (l *Logger) Search(user, action, keyword string, limit int) ([]Entry, error) {
	dates, err := l.ListDates()
	if err != nil {
		return nil, err
	}

	var results []Entry
	for _, date := range dates {
		entries, err := l.ReadDate(date)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if user != "" && !strings.Contains(strings.ToLower(e.User), strings.ToLower(user)) {
				continue
			}
			if action != "" && !strings.Contains(strings.ToLower(e.Action), strings.ToLower(action)) {
				continue
			}
			if keyword != "" {
				lowerKw := strings.ToLower(keyword)
				if !strings.Contains(strings.ToLower(e.Path), lowerKw) &&
					!strings.Contains(strings.ToLower(e.Detail), lowerKw) {
					continue
				}
			}
			results = append(results, e)
			if limit > 0 && len(results) >= limit {
				return results, nil
			}
		}
	}
	return results, nil
}
