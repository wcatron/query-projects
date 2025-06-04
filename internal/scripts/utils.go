package scripts

import "fmt"

func ScriptPathFmt(scriptPath string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", scriptPath)
}
