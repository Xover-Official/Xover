package ai

import (
	"encoding/json"
	"os"
	"sync"
)

type MemoryEntry struct {
	ResourceID string `json:"resource_id"`
	Action     string `json:"action"`
	Reasoning  string `json:"reasoning"`
	Timestamp  string `json:"timestamp"`
}

type ProjectMemory struct {
	Entries []MemoryEntry `json:"entries"`
	mu      sync.Mutex
	path    string
}

func NewProjectMemory(path string) *ProjectMemory {
	m := &ProjectMemory{path: path}
	m.Load()
	return m
}

func (m *ProjectMemory) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.Entries)
}

func (m *ProjectMemory) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(m.Entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

func (m *ProjectMemory) AddEntry(entry MemoryEntry) {
	m.mu.Lock()
	m.Entries = append(m.Entries, entry)
	m.mu.Unlock()
	m.Save()
}

func (m *ProjectMemory) GetDecisionsForResource(id string) []MemoryEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var matched []MemoryEntry
	for _, e := range m.Entries {
		if e.ResourceID == id {
			matched = append(matched, e)
		}
	}
	return matched
}
