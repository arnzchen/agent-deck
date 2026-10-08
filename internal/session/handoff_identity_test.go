package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandoffLauncherIdentityUsesVerifiedSend(t *testing.T) {
	home := identityTestEnv(t)
	dir := filepath.Join(home, ".agent-deck")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"),
		[]byte("[codex]\ncommand='/installed/codex-ad'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ClearUserConfigCache()
	inst := identityTestInstance("codex")
	inst.ProjectPath = t.TempDir()
	got := inst.BuildIdentityPrompt()
	if !strings.Contains(got, "agent-handoff send --key KEY") {
		t.Fatalf("handoff launcher must advertise verified sends:\n%s", got)
	}
	if strings.Contains(got, "agent-deck session send <id-or-title>") {
		t.Fatal("must not advertise conflicting raw send default")
	}
	inst.Tool = "claude"
	got = inst.BuildIdentityPrompt()
	if strings.Contains(got, "agent-handoff send --key KEY") ||
		!strings.Contains(got, "agent-deck session send <id-or-title>") {
		t.Fatal("non-Codex sessions must retain their ordinary messaging instructions")
	}
}
