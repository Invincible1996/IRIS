package root

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/versenilvis/iris/integration"
	"github.com/versenilvis/iris/spec"
	"github.com/versenilvis/iris/spec/alias"
)

func TestMergeResults(t *testing.T) {

	t.Run("Dedup exact match", func(t *testing.T) {
		// Mock history items that might conflict with specs
		res := MergeResults("git", "spec")
		seen := make(map[string]bool)
		for _, r := range res {
			if seen[r.Cmd] {
				t.Errorf("Duplicate suggestion found: %q", r.Cmd)
			}
			seen[r.Cmd] = true
		}
	})

	t.Run("Limit 100", func(t *testing.T) {
		res := MergeResults("a", "history")
		if len(res) > 100 {
			t.Errorf("Expected max 100 suggestions, got %d", len(res))
		}
	})

	t.Run("AI Suggestion Promotion", func(t *testing.T) {
		aiSugg := &spec.Suggestion{
			Cmd:        "git commit -m \"fix(auth): login bug\"",
			Desc:       "AI suggestion",
			Source:     "ai",
			Confidence: 85,
		}
		SetCurrentAISuggestion(aiSugg)
		defer SetCurrentAISuggestion(nil)

		res := MergeResults("git c", "history")
		if len(res) == 0 {
			t.Fatalf("Expected suggestions, got 0")
		}
		if res[0].Cmd != aiSugg.Cmd {
			t.Errorf("Expected AI suggestion at index 0, got %q (confidence %d)", res[0].Cmd, res[0].Confidence)
		}
		if res[0].Source != "ai" {
			t.Errorf("Expected promoted source 'ai', got %q", res[0].Source)
		}
	})

	t.Run("Tool Alias Preserved When Matching Query", func(t *testing.T) {
		alias.Register(&mockGitProvider{})
		defer alias.Reset()
		res := MergeResults("git co", "default")
		foundCo := false
		for _, r := range res {
			if r.Cmd == "git co" {
				foundCo = true
				break
			}
		}
		if !foundCo {
			t.Errorf("expected 'git co' alias suggestion in results for query 'git co', got %v", res)
		}
	})
}

func TestMergeResultsIncludesCDHistoryInSpecMode(t *testing.T) {
	originalSearchHistory := searchHistory
	originalCWD := spec.GetCWD()
	t.Cleanup(func() {
		searchHistory = originalSearchHistory
		spec.SetCWD(originalCWD)
	})

	cwd := t.TempDir()
	for _, name := range []string{"Documents", "Downloads"} {
		if err := os.Mkdir(filepath.Join(cwd, name), 0o755); err != nil {
			t.Fatalf("create test directory %q: %v", name, err)
		}
	}
	spec.SetCWD(cwd)

	searchHistory = func(query string, aliases map[string]string) ([]integration.HistResult, error) {
		if query != "cd " {
			t.Errorf("expected history query %q, got %q", "cd ", query)
		}
		return []integration.HistResult{
			{Cmd: "cd Documents/", Source: "history"},
			{Cmd: "cd Developer/project", Source: "history"},
			{Cmd: "cd Workspace/one", Source: "history"},
			{Cmd: "cd Workspace/two", Source: "history"},
			{Cmd: "cd Workspace/three", Source: "history"},
			{Cmd: "echo cd Documents", Source: "history"},
			{Cmd: "npm run dev", Source: "history"},
			{Cmd: "cdx-not-a-directory", Source: "history"},
		}, nil
	}

	results := MergeResults("cd ", "spec")
	wanted := map[string]bool{
		"cd Developer/project": false,
		"cd Downloads/":        false,
	}
	documentsCount := 0
	visibleDirectory := false
	for i, result := range results {
		if _, ok := wanted[result.Cmd]; ok {
			wanted[result.Cmd] = true
		}
		if i < integration.MaxVisibleItems && result.Desc == "directory" {
			visibleDirectory = true
		}
		if result.Cmd == "cd Documents/" {
			documentsCount++
		}
		if result.Cmd == "echo cd Documents" || result.Cmd == "npm run dev" {
			t.Errorf("unexpected non-cd history suggestion %q", result.Cmd)
		}
		if result.Cmd == "cdx-not-a-directory" {
			t.Errorf("unexpected non-boundary history suggestion %q", result.Cmd)
		}
	}
	for cmd, found := range wanted {
		if !found {
			t.Errorf("expected suggestion %q, got %v", cmd, results)
		}
	}
	if documentsCount != 1 {
		t.Errorf("expected duplicate cd suggestion once, got %d in %v", documentsCount, results)
	}
	if !visibleDirectory {
		t.Errorf("expected a directory in the first %d suggestions, got %v", integration.MaxVisibleItems, results)
	}
}

func TestMergeResultsPromotesPrefixHistoryInSpecMode(t *testing.T) {
	originalSearchHistory := searchHistory
	t.Cleanup(func() {
		searchHistory = originalSearchHistory
	})

	searchHistory = func(query string, aliases map[string]string) ([]integration.HistResult, error) {
		if query != "iris-test-command " {
			t.Errorf("expected history query %q, got %q", "iris-test-command ", query)
		}
		return []integration.HistResult{
			{Cmd: "iris-test-command build", Source: "session"},
			{Cmd: "dart run iris-test-command build", Source: "history"},
		}, nil
	}

	results := MergeResults("iris-test-command ", "spec")
	if len(results) == 0 {
		t.Fatal("expected suggestions")
	}
	if results[0].Cmd != "iris-test-command build" {
		t.Fatalf("expected prefix history first, got %q in %v", results[0].Cmd, results)
	}
	if results[0].Source != "history" {
		t.Errorf("expected hybrid history source, got %q", results[0].Source)
	}
	overlay := integration.NewOverlay()
	overlay.SetQueryAndItems("iris-test-command ", results)
	if ghost := overlay.GetGhostText("iris-test-command ", true); ghost != "build" {
		t.Errorf("expected history ghost text %q, got %q", "build", ghost)
	}
	for _, result := range results {
		if result.Cmd == "dart run iris-test-command build" {
			t.Errorf("unexpected non-prefix history suggestion %q", result.Cmd)
		}
	}
}

func TestMergeResultsEmptyHistoryUsesMostRecentCommand(t *testing.T) {
	originalSearchHistory := searchHistory
	t.Cleanup(func() { searchHistory = originalSearchHistory })

	searchHistory = func(query string, aliases map[string]string) ([]integration.HistResult, error) {
		if query != "" {
			t.Errorf("expected empty history query, got %q", query)
		}
		return []integration.HistResult{
			{Cmd: "git status", Source: "session"},
			{Cmd: "cd project", Source: "session"},
		}, nil
	}

	results := MergeResults("", "history")
	if len(results) != 2 {
		t.Fatalf("expected only history results, got %v", results)
	}
	if results[0].Cmd != "git status" {
		t.Fatalf("expected most recent command first, got %q in %v", results[0].Cmd, results)
	}
}

func TestPrevRecordedCommandState(t *testing.T) {
	origRegistry := spec.Registry
	t.Cleanup(func() {
		spec.Registry = origRegistry
	})

	spec.ResetRegistry()
	spec.Register(&spec.Spec{
		Name: "git",
		Subcommands: []spec.Subcommand{
			{Name: "checkout"},
		},
	})

	setPrevRecordedInfo("git checkout feature-abc", "/repo/dir")
	skel, cwd := getPrevRecordedInfo()
	if skel != "git checkout" || cwd != "/repo/dir" {
		t.Errorf("expected skel='git checkout' and cwd='/repo/dir', got skel=%q, cwd=%q", skel, cwd)
	}
	if gotSkel := getPrevSkeleton(); gotSkel != "git checkout" {
		t.Errorf("expected getPrevSkeleton()='git checkout', got %q", gotSkel)
	}

	setPrevRecordedInfo("", "")
	if gotSkel := getPrevSkeleton(); gotSkel != "" {
		t.Errorf("expected empty getPrevSkeleton() after reset, got %q", gotSkel)
	}
}
