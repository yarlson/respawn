package state

// RunState represents the persistent state of a run to support resuming.
type RunState struct {
	RunID               string `json:"run_id"`
	ActiveTaskID        string `json:"active_task_id"`
	Cycle               int    `json:"cycle"`
	Attempt             int    `json:"attempt"`
	BackendName         string `json:"backend_name"`
	BackendSessionID    string `json:"backend_session_id"`
	LastSavepointCommit string `json:"last_savepoint_commit"`
	ArtifactRootPath    string `json:"artifact_root_path"`
}
