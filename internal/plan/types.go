package plan

import (
	"github.com/wcatron/query-projects/internal/projects"
	lua "github.com/yuin/gopher-lua"
)

// RepoContext represents an execution context for one repo
type RepoContext struct {
	Project *projects.Project
	VM      *lua.LState
}
