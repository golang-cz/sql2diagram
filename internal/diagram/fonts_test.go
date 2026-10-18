package diagram

import (
	"strings"
	"testing"
)

const diagramHash = "d2-1504133605"

// EmbedFonts emits each font-family rule directly followed by its @font-face
// block; the woff blob is a long base64 run inside the src url.
func d2StyleBlock(rule, style string) string {
	return "\n" + rule + "\n@font-face {\n\tfont-family: " + diagramHash + "-font-" + style +
		";\n\tsrc: url(\"data:application/font-woff;base64," + strings.Repeat("QUFB", 8) + ");\n}"
}

func TestStripEmbeddedFonts(t *testing.T) {
	input := `<style type="text/css"><![CDATA[` +
		d2StyleBlock("."+diagramHash+" .text {\n\tfont-family: \""+diagramHash+"-font-regular\";\n}", "regular") +
		d2StyleBlock("."+diagramHash+" .text-bold {\n\tfont-family: \""+diagramHash+"-font-bold\";\n}", "bold") +
		d2StyleBlock("."+diagramHash+" .text-italic {\n\tfont-family: \""+diagramHash+"-font-italic\";\n}", "italic") +
		d2StyleBlock("."+diagramHash+" .text-mono {\n\tfont-family: \""+diagramHash+"-font-mono\";\n}", "mono") +
		d2StyleBlock("."+diagramHash+" .text-mono-bold {\n\tfont-family: \""+diagramHash+"-font-mono-bold\";\n}", "mono-bold") +
		"\n]]></style><g id=\"a\"><text class=\"text fill-N2\">id</text></g>"

	got := string(stripEmbeddedFonts([]byte(input)))

	for _, unwanted := range []string{"@font-face", "font-woff", diagramHash + "-font"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("output still contains %q:\n%s", unwanted, got)
		}
	}

	if strings.Contains(got, "data:application") {
		t.Fatalf("output still contains the font data url:\n%s", got)
	}

	if want := "font-family: " + sansFontStack + ";"; strings.Count(got, want) != 3 {
		t.Fatalf("want 3 sans font-family declarations, got %d:\n%s", strings.Count(got, want), got)
	}
	if want := "font-family: " + monoFontStack + ";"; strings.Count(got, want) != 2 {
		t.Fatalf("want 2 mono font-family declarations, got %d:\n%s", strings.Count(got, want), got)
	}

	if !strings.Contains(got, `<text class="text fill-N2">id</text>`) {
		t.Fatalf("non-font SVG content was altered:\n%s", got)
	}

	if again := string(stripEmbeddedFonts([]byte(got))); again != got {
		t.Fatalf("stripping is not idempotent:\nbefore: %s\nafter:  %s", got, again)
	}
}

func TestStripEmbeddedFontsWithoutFonts(t *testing.T) {
	input := []byte(`<svg><style type="text/css"><![CDATA[
.d2-1504133605 .text {
	font-family: Helvetica;
}
]]></style><text>schema</text></svg>`)

	if got := stripEmbeddedFonts(input); string(got) != string(input) {
		t.Fatalf("SVG without embedded fonts was modified:\nbefore: %s\nafter:  %s", input, got)
	}
}
