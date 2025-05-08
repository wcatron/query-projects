package plan

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wcatron/query-projects/internal/projects"
	"github.com/wcatron/query-projects/internal/scripts"
	lua "github.com/yuin/gopher-lua"
)

// RunFunc creates a Lua function that runs a script in a project
func RunFunc(project *projects.Project) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		script := L.CheckString(1)
		scriptInfo, err := scripts.GetScriptInfo(script)
		if err != nil {
			L.RaiseError("failed to get script info: %v", err)
			return 0
		}
		output, err := scripts.RunScriptForProject(scriptInfo, project.Path, false)
		if err != nil {
			L.RaiseError("failed to run script: %v", err)
			return 0
		}
		L.Push(lua.LString(output.StdoutText))
		return 1
	}
}

// ValueFunc creates a Lua function that reads values from files in a project
func ValueFunc(project *projects.Project) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		file := L.CheckString(1)
		field := L.CheckString(2)

		data, err := os.ReadFile(filepath.Join(project.Path, file))
		if err != nil {
			L.RaiseError("failed to read file %s: %v", file, err)
			return 0
		}

		var value string
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".json":
			var jsonData map[string]interface{}
			if err := json.Unmarshal(data, &jsonData); err != nil {
				L.RaiseError("failed to parse JSON file %s: %v", file, err)
				return 0
			}
			if val, ok := jsonData[field]; ok {
				value = fmt.Sprint(val)
			}
		case ".xml":
			var xmlData map[string]interface{}
			decoder := xml.NewDecoder(bytes.NewReader(data))
			if err := decoder.Decode(&xmlData); err != nil {
				L.RaiseError("failed to parse XML file %s: %v", file, err)
				return 0
			}
			if val, ok := xmlData[field]; ok {
				value = fmt.Sprint(val)
			}
		default:
			L.RaiseError("unsupported file type: %s", ext)
			return 0
		}

		if value == "" {
			L.Push(lua.LString(""))
			return 1
		}

		L.Push(lua.LString(value))
		return 1
	}
}
