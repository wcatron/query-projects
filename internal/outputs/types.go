package outputs

// Result represents the output of running a script on a project
type Result struct {
	ProjectPath string
	Status      string
	StdoutText  string
	StderrText  string
	Index       int
}

// ScriptInfo represents information about a script
type ScriptInfo struct {
	Path    string   `json:"path"`
	Version string   `json:"version"`
	Output  string   `json:"output"`
	Columns []string `json:"columns"`
}
