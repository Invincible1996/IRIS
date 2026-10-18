package sys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/versenilvis/iris/spec"
)

func TestADBInstallSuggestsAPKWithUnicodePrefix(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "唯寻网络.apk"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	oldCWD := spec.GetCWD()
	spec.SetCWD(tmp)
	t.Cleanup(func() { spec.SetCWD(oldCWD) })

	results := spec.Lookup("adb install 唯寻")
	for _, result := range results {
		if result.Cmd == "adb install 唯寻网络.apk" {
			return
		}
	}
	t.Fatalf("expected APK suggestion for Unicode prefix, got %v", results)
}
