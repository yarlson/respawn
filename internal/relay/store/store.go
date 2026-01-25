package relaystore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	relaystore "github.com/yarlson/relay/store"
)

type FileStore struct {
	repoRoot string
	mu       sync.Mutex
}

func New(repoRoot string) *FileStore {
	return &FileStore{repoRoot: repoRoot}
}

func (s *FileStore) SaveWorkflowState(ctx context.Context, state *relaystore.WorkflowState) error {
	_ = ctx
	if state == nil {
		return fmt.Errorf("workflow state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	states, err := s.loadWorkflowStates()
	if err != nil {
		return err
	}
	states[state.WorkflowID] = state

	return s.writeJSONFile(s.statePath(), states)
}

func (s *FileStore) LoadWorkflowState(ctx context.Context, workflowID string) (*relaystore.WorkflowState, error) {
	_ = ctx

	s.mu.Lock()
	defer s.mu.Unlock()

	states, err := s.loadWorkflowStates()
	if err != nil {
		return nil, err
	}
	state, ok := states[workflowID]
	if !ok {
		return nil, relaystore.ErrNotFound
	}
	return state, nil
}

func (s *FileStore) SaveStepSummary(ctx context.Context, summary *relaystore.StepSummary) error {
	_ = ctx
	if summary == nil {
		return fmt.Errorf("step summary is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	summaries, err := s.loadStepSummaries()
	if err != nil {
		return err
	}
	summaries[summary.StepID] = summary

	return s.writeJSONFile(s.stepsPath(), summaries)
}

func (s *FileStore) AppendEvent(ctx context.Context, workflowID string, event *relaystore.Event) error {
	_ = ctx
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	path := s.eventsPath(workflowID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create events dir: %w", err)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open events file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("append event: %w", err)
	}

	return nil
}

func (s *FileStore) StreamRawPayload(ctx context.Context, workflowID, stepID string, data []byte) error {
	_ = ctx
	path := s.rawPath(workflowID, stepID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create raw dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open raw file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("append raw payload: %w", err)
	}

	if len(data) > 0 && data[len(data)-1] != '\n' {
		if _, err := f.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("append raw newline: %w", err)
		}
	}

	return nil
}

func (s *FileStore) statePath() string {
	return filepath.Join(s.repoRoot, ".turbine", "state", "relay.json")
}

func (s *FileStore) stepsPath() string {
	return filepath.Join(s.repoRoot, ".turbine", "state", "steps.json")
}

func (s *FileStore) eventsPath(workflowID string) string {
	return filepath.Join(s.repoRoot, ".turbine", "runs", workflowID, "events.log")
}

func (s *FileStore) rawPath(workflowID, stepID string) string {
	return filepath.Join(s.repoRoot, ".turbine", "runs", workflowID, "raw", stepID+".ndjson")
}

func (s *FileStore) loadWorkflowStates() (map[string]*relaystore.WorkflowState, error) {
	var states map[string]*relaystore.WorkflowState
	if err := s.readJSONFile(s.statePath(), &states); err != nil {
		if os.IsNotExist(err) {
			return map[string]*relaystore.WorkflowState{}, nil
		}
		return nil, err
	}
	if states == nil {
		states = map[string]*relaystore.WorkflowState{}
	}
	return states, nil
}

func (s *FileStore) loadStepSummaries() (map[string]*relaystore.StepSummary, error) {
	var summaries map[string]*relaystore.StepSummary
	if err := s.readJSONFile(s.stepsPath(), &summaries); err != nil {
		if os.IsNotExist(err) {
			return map[string]*relaystore.StepSummary{}, nil
		}
		return nil, err
	}
	if summaries == nil {
		summaries = map[string]*relaystore.StepSummary{}
	}
	return summaries, nil
}

func (s *FileStore) readJSONFile(path string, dest interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *FileStore) writeJSONFile(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0644); err != nil {
		return fmt.Errorf("write json file: %w", err)
	}
	return nil
}
