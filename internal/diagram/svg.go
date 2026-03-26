package diagram

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

const (
	pkRowBackgroundColor = "#D3F9D8"
	uniqueRowColor       = "#FFF3BF"
	fkRowBackgroundColor = "#D0EBFF"
)

var (
	tableGroupRegex = regexp.MustCompile(`(?s)<g id="[^"]+"><g class="shape" >.*?</g></g>`)
	tableRectRegex  = regexp.MustCompile(`<rect x="([^"]+)" y="([^"]+)" width="([^"]+)" height="([^"]+)" class="shape stroke-N1 fill-N7"[^>]*/>`)
	headerRectRegex = regexp.MustCompile(`<rect x="[^"]+" y="[^"]+" width="[^"]+" height="36\.000000" class="class_header fill-N1" />`)
	rowMetaRegex    = regexp.MustCompile(`(?s)<text x="[^"]+" y="[^"]+" class="text fill-N2"[^>]*>([^<]*)</text><text x="[^"]+" y="[^"]+" class="text fill-AA2"[^>]*/><line x1="[^"]+" x2="[^"]+" y1="([^"]+)" y2="[^"]+" class=" stroke-N1"`)
	uniqueRowKeyRE = regexp.MustCompile(`__uq_`)
)

func postProcessSVG(svg []byte) []byte {
	result := tableGroupRegex.ReplaceAllStringFunc(string(svg), func(group string) string {
		if !strings.Contains(group, "class_header") {
			return group
		}

		tableRectMatch := tableRectRegex.FindStringSubmatch(group)
		if len(tableRectMatch) < 4 {
			return group
		}

		headerRect := headerRectRegex.FindStringIndex(group)
		if len(headerRect) != 2 {
			return group
		}

		tableX, err := strconv.ParseFloat(tableRectMatch[1], 64)
		if err != nil {
			return group
		}
		tableWidth, err := strconv.ParseFloat(tableRectMatch[3], 64)
		if err != nil {
			return group
		}

		// Inset by border stroke width so backgrounds don't cover table borders.
		const borderWidth = 2
		bgX := tableX + borderWidth
		bgWidth := tableWidth - borderWidth*2

		var backgrounds []string
		rowMatches := rowMetaRegex.FindAllStringSubmatch(group, -1)
		for _, row := range rowMatches {
			if len(row) < 3 {
				continue
			}

			rowLabel := html.UnescapeString(row[1])
			lineY, err := strconv.ParseFloat(row[2], 64)
			if err != nil {
				continue
			}

			color := rowBackgroundByConstraint(rowLabel)
			if color == "" {
				continue
			}

			backgrounds = append(backgrounds, fmt.Sprintf(
				`<rect x="%.6f" y="%.6f" width="%.6f" height="36.000000" class="row_constraint_bg" style="fill:%s;stroke:none;" />`,
				bgX, lineY-36, bgWidth, color,
			))
		}

		if len(backgrounds) == 0 {
			return group
		}

		insertAt := headerRect[1]
		return group[:insertAt] + strings.Join(backgrounds, "") + group[insertAt:]
	})

	result = uniqueRowKeyRE.ReplaceAllString(result, "")

	return []byte(result)
}

func rowBackgroundByConstraint(label string) string {
	switch {
	case strings.Contains(label, "(FK)"):
		return fkRowBackgroundColor
	case strings.Contains(label, "(PK)"):
		return pkRowBackgroundColor
	case strings.Contains(label, "(UNIQUE)"):
		return uniqueRowColor
	default:
		return ""
	}
}

