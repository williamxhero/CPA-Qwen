package plugin

import (
	"strconv"
	"strings"
	"testing"
)

// The repository ships its own semantic version. Forgetting to re-run
// ops/update-versions.ps1 after a change must fail here instead of leaking a
// stale or hand-edited identity.
func TestGeneratedVersionShape(t *testing.T) {
	if !isSemverTriple(generatedVersion) {
		t.Fatalf("generatedVersion = %q, want major.minor.patch", generatedVersion)
	}
}

func isSemverTriple(value string) bool {
	segments := strings.Split(value, ".")
	if len(segments) != 3 {
		return false
	}
	for _, segment := range segments {
		if segment == "" || strings.HasPrefix(segment, "-") || (len(segment) > 1 && segment[0] == '0') {
			return false
		}
		if _, err := strconv.Atoi(segment); err != nil {
			return false
		}
	}
	return true
}
