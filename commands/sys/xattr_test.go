package sys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/versenilvis/iris/spec"
)

func TestXattrSuggestsAbsolutePaths(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"Screenflare.app", "Shottr.app"} {
		if err := os.Mkdir(filepath.Join(tmp, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	results := spec.Lookup("xattr -cr " + filepath.Join(tmp, "S"))
	want := map[string]bool{
		"xattr -cr " + filepath.Join(tmp, "Screenflare.app") + "/": false,
		"xattr -cr " + filepath.Join(tmp, "Shottr.app") + "/":      false,
	}
	for _, result := range results {
		if _, ok := want[result.Cmd]; ok {
			want[result.Cmd] = true
		}
	}
	for cmd, found := range want {
		if !found {
			t.Errorf("expected absolute path suggestion %q, got %v", cmd, results)
		}
	}
}
