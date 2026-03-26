package diagram

import (
	"strings"
	"testing"
)

func TestPostProcessSVGAddsConstraintBackgrounds(t *testing.T) {
	t.Parallel()

	input := `<g id="demo"><g class="shape" ><rect x="0.000000" y="0.000000" width="200.000000" height="180.000000" class="shape stroke-N1 fill-N7" style="stroke-width:2;" /><rect x="0.000000" y="0.000000" width="200.000000" height="36.000000" class="class_header fill-N1" /><text x="10.000000" y="59.000000" class="text fill-B2" style="text-anchor:start;font-size:20px">id</text><text x="100.000000" y="59.000000" class="text fill-N2" style="text-anchor:start;font-size:20px">uuid (PK)</text><text x="190.000000" y="59.000000" class="text fill-AA2" style="text-anchor:end;font-size:20px;letter-spacing:2px" /><line x1="0.000000" x2="200.000000" y1="72.000000" y2="72.000000" class=" stroke-N1" style="stroke-width:2" /><text x="10.000000" y="95.000000" class="text fill-B2" style="text-anchor:start;font-size:20px">org_id</text><text x="100.000000" y="95.000000" class="text fill-N2" style="text-anchor:start;font-size:20px">uuid (FK)</text><text x="190.000000" y="95.000000" class="text fill-AA2" style="text-anchor:end;font-size:20px;letter-spacing:2px" /><line x1="0.000000" x2="200.000000" y1="108.000000" y2="108.000000" class=" stroke-N1" style="stroke-width:2" /><text x="10.000000" y="131.000000" class="text fill-B2" style="text-anchor:start;font-size:20px">email</text><text x="100.000000" y="131.000000" class="text fill-N2" style="text-anchor:start;font-size:20px">text (UNIQUE)</text><text x="190.000000" y="131.000000" class="text fill-AA2" style="text-anchor:end;font-size:20px;letter-spacing:2px" /><line x1="0.000000" x2="200.000000" y1="144.000000" y2="144.000000" class=" stroke-N1" style="stroke-width:2" /></g></g>`

	output := string(postProcessSVG([]byte(input)))

	if count := strings.Count(output, `class="row_constraint_bg"`); count != 3 {
		t.Fatalf("expected 3 constraint backgrounds, got %d: %s", count, output)
	}

	if !strings.Contains(output, pkRowBackgroundColor) {
		t.Fatalf("expected PK row color in SVG: %s", output)
	}

	if !strings.Contains(output, fkRowBackgroundColor) {
		t.Fatalf("expected FK row color in SVG: %s", output)
	}

	if !strings.Contains(output, uniqueRowColor) {
		t.Fatalf("expected UNIQUE row color in SVG: %s", output)
	}
}

func TestPostProcessSVGStripsUniqueKeyPrefixAndPreservesRealColumns(t *testing.T) {
	t.Parallel()

	input := `<g id="demo"><g class="shape" ><rect x="0.000000" y="0.000000" width="345.000000" height="180.000000" class="shape stroke-N1 fill-N7" style="stroke-width:2;" /><rect x="0.000000" y="0.000000" width="345.000000" height="36.000000" class="class_header fill-N1" /><text x="10.000000" y="59.000000" class="text fill-B2" style="text-anchor:start;font-size:20px">unique_1</text><text x="242.000000" y="59.000000" class="text fill-N2" style="text-anchor:start;font-size:20px">text NULL</text><text x="325.000000" y="59.000000" class="text fill-AA2" style="text-anchor:end;font-size:20px;letter-spacing:2px" /><line x1="0.000000" x2="345.000000" y1="72.000000" y2="72.000000" class=" stroke-N1" style="stroke-width:2" /><text x="10.000000" y="95.000000" class="text fill-B2" style="text-anchor:start;font-size:20px">__uq_(a, b)</text><text x="242.000000" y="95.000000" class="text fill-N2" style="text-anchor:start;font-size:20px">(UNIQUE)</text><text x="325.000000" y="95.000000" class="text fill-AA2" style="text-anchor:end;font-size:20px;letter-spacing:2px" /><line x1="0.000000" x2="345.000000" y1="108.000000" y2="108.000000" class=" stroke-N1" style="stroke-width:2" /></g></g>`

	output := string(postProcessSVG([]byte(input)))

	if strings.Contains(output, "__uq_") {
		t.Fatalf("expected __uq_ prefix to be stripped from SVG: %s", output)
	}

	if !strings.Contains(output, `>(a, b)</text>`) {
		t.Fatalf("expected column list to remain after prefix strip: %s", output)
	}

	if !strings.Contains(output, `>unique_1</text>`) {
		t.Fatalf("expected real unique_1 column label to remain intact: %s", output)
	}
}

func TestRowBackgroundByConstraintPrefersForeignKeyColorOverPrimaryKey(t *testing.T) {
	t.Parallel()

	color := rowBackgroundByConstraint("uuid (PK) (FK)")
	if color != fkRowBackgroundColor {
		t.Fatalf("expected FK color %q for combined PK/FK label, got %q", fkRowBackgroundColor, color)
	}
}
