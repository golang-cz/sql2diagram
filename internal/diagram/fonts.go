package diagram

import (
	"fmt"
	"regexp"
)

const (
	sansFontStack = `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif`
	monoFontStack = `ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`
)

var (
	// d2's EmbedFonts emits @font-face blocks with no nested braces.
	fontFaceRegex = regexp.MustCompile(`(?s)\n?@font-face \{[^}]*\}`)
	// d2 names its embedded families d2-<hash>-font-<style>; the mono styles
	// cover Source Code Pro, the rest Source Sans Pro.
	fontFamilyRegex = regexp.MustCompile(`font-family: "d2-\d+-font-(mono(?:-bold|-italic)?|semibold|bold|italic|regular)";`)
	monoFontRegex   = regexp.MustCompile(`^mono`)
)

// stripEmbeddedFonts removes the woff fonts d2 embeds in the SVG and points
// the CSS rules at system font stacks instead. Embedded fonts carry zlib
// streams that vary between Go toolchain versions, so the same schema would
// otherwise render differently on different machines.
func stripEmbeddedFonts(svg []byte) []byte {
	out := fontFaceRegex.ReplaceAll(svg, nil)
	return []byte(fontFamilyRegex.ReplaceAllStringFunc(string(out), func(decl string) string {
		stack := sansFontStack
		style := fontFamilyRegex.FindStringSubmatch(decl)[1]
		if monoFontRegex.MatchString(style) {
			stack = monoFontStack
		}
		return fmt.Sprintf("font-family: %s;", stack)
	}))
}
