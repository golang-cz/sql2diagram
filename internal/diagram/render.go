package diagram

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-cz/sql2diagram/internal/schema"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

func Render(ctx context.Context, schemaDef *schema.Schema) ([]byte, error) {
	_, graph, err := d2lib.Compile(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("d2 compile: %w", err)
	}

	graph, err = transformGraph(schemaDef, graph)
	if err != nil {
		return nil, fmt.Errorf("transform graph: %w", err)
	}

	// direction: up makes dagre place FK targets (parent tables) at the top.
	script := "direction: up\n" + d2format.Format(graph.AST)

	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("d2 textmeasure new ruler: %w", err)
	}

	diagram, _, err := d2lib.Compile(ctx, script, &d2lib.CompileOptions{
		Layout: d2dagrelayout.DefaultLayout,
		Ruler:  ruler,
	})
	if err != nil {
		return nil, fmt.Errorf("d2 compile new ruler: %w", err)
	}

	out, err := d2svg.Render(diagram, &d2svg.RenderOpts{Pad: d2svg.DEFAULT_PADDING})
	if err != nil {
		return nil, fmt.Errorf("d2 render to svg: %w", err)
	}

	return postProcessSVG(out), nil
}

func transformGraph(schemaDef *schema.Schema, g *d2graph.Graph) (*d2graph.Graph, error) {
	for _, table := range schemaDef.Tables {
		_, newKey, err := d2oracle.Create(g, table.Name)
		if err != nil {
			return nil, fmt.Errorf("d2 oracle create: %w", err)
		}

		shape := "sql_table"
		if _, err = d2oracle.Set(g, fmt.Sprintf("%s.shape", newKey), nil, &shape); err != nil {
			return nil, fmt.Errorf("d2 oracle create: %w", err)
		}

		for _, column := range table.Columns {
			columnType := column.Type
			if column.Length > 0 {
				columnType = fmt.Sprintf("%s(%d)", column.Type, column.Length)
			}

			if !column.Constraints.Has("not null") {
				columnType += " NULL"
			}

			if column.Constraints.Has("primary") {
				columnType += " (PK)"
			}
			if column.Constraints.Has("unique") {
				columnType += " (UNIQUE)"
			}
			if len(column.ForeignKeyReferences) > 0 {
				columnType += " (FK)"
			}

			if _, err = d2oracle.Set(g, fmt.Sprintf("%s.%s", table.Name, column.Name), nil, &columnType); err != nil {
				return nil, fmt.Errorf("d2 set: %w", err)
			}

			for _, fk := range column.ForeignKeyReferences {
				ref := fmt.Sprintf("%s.%s -> %s.%s", table.Name, column.Name, fk.Table, fk.Column)
				if _, _, err = d2oracle.Create(g, ref); err != nil {
					return nil, fmt.Errorf("d2 oracle create: %w", err)
				}
			}
		}

		for _, uniqueColumns := range table.UniqueConstraints {
			rowName := "__uq_(" + strings.Join(uniqueColumns, ", ") + ")"
			rowType := "(UNIQUE)"

			if _, err = d2oracle.Set(g, fmt.Sprintf("%s.%s", table.Name, rowName), nil, &rowType); err != nil {
				return nil, fmt.Errorf("d2 set unique constraint: %w", err)
			}
		}
	}

	return g, nil
}
