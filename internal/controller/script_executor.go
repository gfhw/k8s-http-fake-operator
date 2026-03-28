package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	httpteststubv1 "httpteststub.example.com/api/v1"
)

type ScriptExecutor struct {
	scriptDir string
}

func NewScriptExecutor() *ScriptExecutor {
	scriptDir := os.Getenv("SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = "/scripts"
	}

	// 确保脚本目录存在
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		log.Printf("Warning: Failed to create script directory: %v", err)
	}

	return &ScriptExecutor{
		scriptDir: scriptDir,
	}
}

func (e *ScriptExecutor) Execute(script *httpteststubv1.Script, requestContext map[string]interface{}) (int, map[string]string, []byte, error) {
	var cmd *exec.Cmd
	scriptPath := e.getScriptPath(script)

	if script.Content != "" {
		scriptPath = e.createTempScript(script)
		defer os.Remove(scriptPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(script.Timeout)*time.Second)
	defer cancel()

	switch script.Type {
	case "shell", "sh", "bash":
		cmd = exec.CommandContext(ctx, "bash", scriptPath)
	case "python", "python3", "py":
		cmd = exec.CommandContext(ctx, "python3", scriptPath)
	default:
		cmd = exec.CommandContext(ctx, scriptPath)
	}

	if len(script.Args) > 0 {
		cmd.Args = append([]string{scriptPath}, script.Args...)
	}

	env := e.buildEnvironment(script, requestContext)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return 504, nil, nil, fmt.Errorf("script execution timeout after %d seconds", script.Timeout)
		}
		return 500, nil, nil, fmt.Errorf("script execution failed: %v, stderr: %s", err, stderr.String())
	}

	return 200, nil, stdout.Bytes(), nil
}

func (e *ScriptExecutor) getScriptPath(script *httpteststubv1.Script) string {
	if filepath.IsAbs(script.Path) {
		return script.Path
	}
	return filepath.Join(e.scriptDir, script.Path)
}

func (e *ScriptExecutor) createTempScript(script *httpteststubv1.Script) string {
	ext := ".sh"
	switch script.Type {
	case "python", "python3", "py":
		ext = ".py"
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("script_%s_*%s", script.Name, ext))
	if err != nil {
		return ""
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(script.Content)
	if err != nil {
		return ""
	}

	return tmpFile.Name()
}

func (e *ScriptExecutor) buildEnvironment(script *httpteststubv1.Script, requestContext map[string]interface{}) []string {
	env := os.Environ()

	for key, value := range script.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	for key, value := range requestContext {
		env = append(env, fmt.Sprintf("REQUEST_%s=%s", strings.ToUpper(key), fmt.Sprintf("%v", value)))
	}

	return env
}

func ParseScriptOutput(output []byte) (interface{}, map[string]string, int) {
	var result interface{}
	var headers map[string]string
	statusCode := 200

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return nil, nil, 500
	}

	if err := json.Unmarshal([]byte(lines[0]), &result); err == nil {
		if len(lines) > 1 {
			if err := json.Unmarshal([]byte(lines[1]), &headers); err != nil {
				headers = make(map[string]string)
			}
		}
		if len(lines) > 2 {
			if err := json.Unmarshal([]byte(lines[2]), &statusCode); err != nil {
				statusCode = 200
			}
		}
	} else {
		result = string(output)
	}

	return result, headers, statusCode
}

func GetRequestContext(c interface{}) map[string]interface{} {
	return map[string]interface{}{
		"method":  "GET",
		"path":    "/",
		"headers": map[string]string{},
		"body":    "",
	}
}

func CopyRequestContext(ctx interface{}) map[string]interface{} {
	return map[string]interface{}{
		"method":  "GET",
		"path":    "/",
		"headers": map[string]string{},
		"body":    "",
	}
}
