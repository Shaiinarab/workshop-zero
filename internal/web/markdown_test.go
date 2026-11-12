package web

import (
	"strings"
	"testing"
)

func TestMDLiteRendersWorkshopMarkdown(t *testing.T) {
	got := string(mdLite("# Guide\n\n1. First step\n2. Second step\n\n> Keep thinking\n\n[Reference](https://example.com)"))
	for _, want := range []string{"<h2>Guide</h2>", "<ol><li>First step</li><li>Second step</li></ol>", "<blockquote><p>Keep thinking</p></blockquote>", `href="https://example.com"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered markdown missing %q: %s", want, got)
		}
	}
}

func TestMDLiteEscapesRawHTML(t *testing.T) {
	got := string(mdLite("<script>alert(1)</script>"))
	if strings.Contains(got, "<script>") || !strings.Contains(got, "&lt;script&gt;") {
		t.Fatalf("raw HTML was not escaped: %s", got)
	}
}
