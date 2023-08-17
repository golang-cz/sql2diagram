package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"flag"
)

type Schema struct {
	Tables []*Table
}

type ForeignReference struct {
	Table  string
	Column string
}

type Table struct {
	Name    string
	Columns []*Column
}

type Column struct {
	Name                 string
	Type                 string
	Constraints          []string
	ForeignKeyReferences []*ForeignReference
	Length               int
}

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("schema", "", "input sql schema file")
)

func main() {
	f, _ := os.ReadFile(filepath.Join("schema.sql"))

	schemaSql := string(f)

	if schemaSql == "" {
		log.Fatal("schema was not provided, file is empty")
	}

	// Parse the SQL statement using pg_query_go
	tree, err := pg_query.Parse(strings.TrimSpace(schemaSql))
	if err != nil {
		fmt.Println("Error parsing SQL statement:", err)
		return
	}

	schema, err := astTreeToSchema(tree)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// Start with a new, empty graph
	_, graph, _ := d2lib.Compile(ctx, "", nil)

	graph = transformGraph(schema, graph)

	// Turn the graph into a script (which would just be "meow")
	script := d2format.Format(graph.AST)

	// Initialize a ruler to measure font glyphs
	ruler, _ := textmeasure.NewRuler()

	// Compile the script into a diagram
	diagram, _, _ := d2lib.Compile(context.Background(), script, &d2lib.CompileOptions{
		Layout: d2dagrelayout.DefaultLayout,
		Ruler:  ruler,
	})

	// Render to SVG
	out, _ := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad: d2svg.DEFAULT_PADDING,
	})

	// Write to disk
	err = os.WriteFile("out.svg", out, 07770)
	if err != nil {
		log.Fatal(err)
	}
}

func transformGraph(schema *Schema, g *d2graph.Graph) *d2graph.Graph {
	for _, table := range schema.Tables {
		// Create an object with the ID set to the table name
		_, newKey, _ := d2oracle.Create(g, table.Name)
		// Set the shape of the newly created object to be D2 shape type "sql_table"
		shape := "sql_table"
		_, _ = d2oracle.Set(g, fmt.Sprintf("%s.shape", newKey), nil, &shape)

		for _, column := range table.Columns {
			_, _ = d2oracle.Set(g, fmt.Sprintf("%s.%s", table.Name, column.Name), nil, &column.Type)

			for _, foreignReference := range column.ForeignKeyReferences {
				_, _, _ = d2oracle.Create(g, fmt.Sprintf(
					"%s.%s -> %s.%s",
					table.Name,
					column.Name,
					foreignReference.Table,
					foreignReference.Column,
				))
			}

		}
	}

	return g
}

func astTreeToSchema(tree *pg_query.ParseResult) (*Schema, error) {
	schema := &Schema{
		Tables: make([]*Table, 0),
	}

	for _, stmt := range tree.Stmts {
		switch node := stmt.Stmt.Node.(type) {
		case *pg_query.Node_CreateStmt:
			schema.Tables = append(schema.Tables, toTable(node))
		case *pg_query.Node_AlterTableStmt:
			err := alterTableStmt(schema, node.AlterTableStmt)
			if err != nil {
				return nil, fmt.Errorf("failed to handle alter table stmt: %w", err)
			}
		}

	}

	return schema, nil
}

func alterTableStmt(schema *Schema, stmt *pg_query.AlterTableStmt) error {
	var sourceTable *Table

	for _, t := range schema.Tables {
		if t.Name == stmt.Relation.Relname {
			sourceTable = t
			break
		}
	}

	if sourceTable == nil {
		return fmt.Errorf("sourceTable could not be found in schema")
	}

	for _, cmd := range stmt.Cmds {
		node, ok := cmd.Node.(*pg_query.Node_AlterTableCmd)
		if !ok {
			continue
		}

		if node.AlterTableCmd.Subtype != pg_query.AlterTableType_AT_AddConstraint {
			continue
		}

		constraint, ok := node.AlterTableCmd.Def.Node.(*pg_query.Node_Constraint)
		if !ok {
			continue
		}

		if constraint.Constraint.Contype != pg_query.ConstrType_CONSTR_FOREIGN {
			continue
		}

		foreignReference := &ForeignReference{
			Table: constraint.Constraint.Pktable.Relname,
		}

		for _, pkattr := range constraint.Constraint.PkAttrs {
			node, ok := pkattr.Node.(*pg_query.Node_String_)
			if !ok {
				continue
			}

			var column *Column
			for _, col := range sourceTable.Columns {
				if col.Name == node.String_.Sval {
					column = col
					break
				}
			}

			foreignReference.Column = node.String_.Sval

			column.ForeignKeyReferences = append(column.ForeignKeyReferences, foreignReference)
		}
	}

	return nil
}

func toTable(stmt interface{}) *Table {
	table := &Table{}

	switch stmt.(type) {
	case *pg_query.Node_CreateStmt:
		tableNodeStmt := stmt.(*pg_query.Node_CreateStmt)
		tableStmt := tableNodeStmt.CreateStmt

		table.Name = tableStmt.Relation.Relname

		for _, columnNode := range tableStmt.TableElts {
			columnNodeDefinition, ok := columnNode.Node.(*pg_query.Node_ColumnDef)
			if !ok {
				continue
			}

			table.Columns = append(table.Columns, generateColumnProperties(columnNodeDefinition.ColumnDef))
		}
	default:
	}

	return table
}

func generateColumnProperties(columnDefinition *pg_query.ColumnDef) *Column {
	column := &Column{
		Name:                 columnDefinition.Colname,
		Constraints:          make([]string, 0),
		ForeignKeyReferences: make([]*ForeignReference, 0),
	}

	for _, node := range columnDefinition.TypeName.Names {
		stringNode, ok := node.Node.(*pg_query.Node_String_)
		if !ok {
			fmt.Printf("unknown name node %v", stringNode)
			continue
		}

		column.Type = stringNode.String_.Sval

		// get length value
		for _, mod := range columnDefinition.TypeName.Typmods {
			aConst, ok := mod.Node.(*pg_query.Node_AConst)
			if !ok {
				continue
			}

			integer, ok := aConst.AConst.Val.(*pg_query.A_Const_Ival)
			if !ok {
				continue
			}

			column.Length = int(integer.Ival.GetIval())
		}

		// map not null constraint to string
		for _, constraint := range columnDefinition.Constraints {
			nodeConstraint, ok := constraint.Node.(*pg_query.Node_Constraint)
			if !ok {
				continue
			}

			if nodeConstraint.Constraint.Contype == pg_query.ConstrType_CONSTR_PRIMARY {
				column.Constraints = append(column.Constraints, "primary")
			}

			if nodeConstraint.Constraint.Contype == pg_query.ConstrType_CONSTR_FOREIGN {
				foreignReference := &ForeignReference{
					Table: nodeConstraint.Constraint.Pktable.Relname,
				}

				found := false

				for _, pkattr := range nodeConstraint.Constraint.PkAttrs {
					node, ok := pkattr.Node.(*pg_query.Node_String_)
					if !ok {
						continue
					}

					for _, fkr := range column.ForeignKeyReferences {
						if fkr.Table == nodeConstraint.Constraint.Pktable.Relname && fkr.Column == node.String_.Sval {
							found = true
						}
					}

					foreignReference.Column = node.String_.Sval
				}

				if !found {
					column.ForeignKeyReferences = append(column.ForeignKeyReferences, foreignReference)
				}
			}

			if nodeConstraint.Constraint.Contype == pg_query.ConstrType_CONSTR_NOTNULL {
				alreadyExists := false
				for _, c := range column.Constraints {
					if c == "not null" {
						alreadyExists = true
					}
				}

				if !alreadyExists {
					column.Constraints = append(column.Constraints, "not null")
				}
			}
		}
	}

	return column
}
