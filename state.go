package main

import (
	"encoding/json"
	"os"
	"time"
)

type State struct {
	LastIP          string       `json:"last_ip"`
	LastChangeTime  time.Time    `json:"last_change_time"`
	NextExpected    time.Time    `json:"next_expected"`
	LastCheckTime   time.Time    `json:"last_check_time"`
	CurrentIP       string       `json:"current_ip"`
	ChangeCount     int          `json:"change_count"`
	LastUpdateLog   []UpdateLog  `json:"last_update_log"`
}

type UpdateLog struct {
	Time    time.Time `json:"time"`
	OldIP   string    `json:"old_ip"`
	NewIP   string    `json:"new_ip"`
	Results []LogEntry `json:"results"`
}

type LogEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // file / command
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func loadState(path string) *State {
	data, err := os.ReadFile(path)
	if err != nil {
		return &State{}
	}
	s := &State{}
	json.Unmarshal(data, s)
	return s
}

func saveState(path string, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *State) HasChanged(newIP string) bool {
	return s.LastIP != "" && s.LastIP != newIP
}

func (s *State) RecordChange(oldIP, newIP string, results []LogEntry) {
	s.LastIP = newIP
	s.CurrentIP = newIP
	s.LastChangeTime = time.Now()
	s.NextExpected = time.Now().Add(5 * 24 * time.Hour)
	s.ChangeCount++
	s.LastUpdateLog = append(s.LastUpdateLog, UpdateLog{
		Time:    time.Now(),
		OldIP:   oldIP,
		NewIP:   newIP,
		Results: results,
	})
	// Only keep last 20 logs
	if len(s.LastUpdateLog) > 20 {
		s.LastUpdateLog = s.LastUpdateLog[len(s.LastUpdateLog)-20:]
	}
}

func (s *State) RecordCheck(ip string) {
	s.CurrentIP = ip
	s.LastCheckTime = time.Now()
	if s.LastIP == "" {
		s.LastIP = ip
	}
}
