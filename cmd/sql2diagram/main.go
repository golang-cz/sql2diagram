package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	pgQuery "github.com/pganalyze/pg_query_go/v6"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/textmeasure"
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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var input []byte
	var err error

	switch len(os.Args) {
	case 1:
		// No arguments - read from stdin
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(fmt.Errorf("read from stdin: %w", err))
		}

	case 2:
		// One argument - read from file
		filePath := os.Args[1]
		input, err = os.ReadFile(filePath)
		if err != nil {
			log.Fatal(fmt.Errorf("read file %s: %w", filePath, err))
		}

	default:
		log.Fatal("Usage: sql2diagram [schema_file] > schema.svg")
	}

	if err := generateDiagram(ctx, string(input)); err != nil {
		log.Fatal(err)
	}
}

func generateDiagram(ctx context.Context, input string) error {
	schemaSQL := strings.TrimSpace(input)

	if schemaSQL == "" {
		return fmt.Errorf("schema was not provided, input is empty")
	}

	// Parse the SQL statement using pg_query_go
	tree, err := pgQuery.Parse(schemaSQL)
	if err != nil {
		return fmt.Errorf("parse SQL statement: %w", err)
	}

	schema, err := astTreeToSchema(tree)
	if err != nil {
		return fmt.Errorf("ast tree to schema: %w", err)
	}

	// Start with a new, empty graph
	_, graph, err := d2lib.Compile(ctx, "", nil)
	if err != nil {
		return fmt.Errorf("d2 compile: %w", err)
	}

	graph, err = transformGraph(schema, graph)
	if err != nil {
		return fmt.Errorf("transform graph: %w", err)
	}

	// Turn the graph into a script (which would just be "meow")
	script := d2format.Format(graph.AST)

	// Initialize a ruler to measure font glyphs
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return fmt.Errorf("d2 textmeasure new ruler: %w", err)
	}

	// Compile the script into a diagram
	diagram, _, err := d2lib.Compile(ctx, script, &d2lib.CompileOptions{
		Layout: d2dagrelayout.DefaultLayout,
		Ruler:  ruler,
	})
	if err != nil {
		return fmt.Errorf("d2 compile new ruler: %w", err)
	}

	// Render to SVG
	out, err := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad: d2svg.DEFAULT_PADDING,
	})
	if err != nil {
		return fmt.Errorf("d2 render to svg: %w", err)
	}

	if _, err = os.Stdout.Write(out); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	return nil
}

func transformGraph(schema *Schema, g *d2graph.Graph) (*d2graph.Graph, error) {
	for _, table := range schema.Tables {
		// Create an object with the ID set to the table name
		_, newKey, err := d2oracle.Create(g, table.Name)
		if err != nil {
			return nil, fmt.Errorf("d2 oracle create: %w", err)
		}

		// Set the shape of the newly created object to be D2 shape type "sql_table"
		shape := "sql_table"
		if _, err = d2oracle.Set(g, fmt.Sprintf("%s.shape", newKey), nil, &shape); err != nil {
			return nil, fmt.Errorf("d2 oracle create: %w", err)
		}

		for _, column := range table.Columns {
			// Use type as the value, add NULL if column can be null
			columnType := column.Type
			if column.Length > 0 {
				columnType = fmt.Sprintf("%s(%d)", column.Type, column.Length)
			}

			// Add NULL if column can be null (doesn't have NOT NULL constraint)
			if !contains(column.Constraints, "not null") {
				columnType += " NULL"
			}

			if contains(column.Constraints, "primary") {
				columnType = fmt.Sprintf("%s (PK)", columnType)
			}

			if len(column.ForeignKeyReferences) > 0 {
				columnType = fmt.Sprintf("%s (FK)", columnType)
			}

			if _, err = d2oracle.Set(g, fmt.Sprintf("%s.%s", table.Name, column.Name), nil, &columnType); err != nil {
				return nil, fmt.Errorf("d2 set: %w", err)
			}

			for _, foreignReference := range column.ForeignKeyReferences {
				// Find the referenced column to get its formatted name
				var referencedColumnName string
				for _, t := range schema.Tables {
					if t.Name == foreignReference.Table {
						for _, col := range t.Columns {
							if col.Name == foreignReference.Column {
								referencedColumnName = col.Name
								break
							}
						}
						break
					}
				}
				if referencedColumnName == "" {
					referencedColumnName = foreignReference.Column
				}

				tableReferences := fmt.Sprintf(
					"%s.%s -> %s.%s",
					table.Name,
					column.Name,
					foreignReference.Table,
					referencedColumnName,
				)
				if _, _, err = d2oracle.Create(g, tableReferences); err != nil {
					return nil, fmt.Errorf("d2 oracle create: %w", err)
				}

			}

		}
	}

	return g, nil
}

func astTreeToSchema(tree *pgQuery.ParseResult) (*Schema, error) {
	schema := &Schema{
		Tables: make([]*Table, 0),
	}

	for _, stmt := range tree.Stmts {
		switch node := stmt.Stmt.Node.(type) {
		case *pgQuery.Node_CreateStmt:
			schema.Tables = append(schema.Tables, toTable(node))
		case *pgQuery.Node_AlterTableStmt:
			if err := alterTableStmt(schema, node.AlterTableStmt); err != nil {
				return nil, fmt.Errorf("handle alter table stmt: %w", err)
			}
		}
	}

	return schema, nil
}

func alterTableStmt(schema *Schema, stmt *pgQuery.AlterTableStmt) error {
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
		node, ok := cmd.Node.(*pgQuery.Node_AlterTableCmd)
		if !ok {
			continue
		}

		if node.AlterTableCmd.Subtype != pgQuery.AlterTableType_AT_AddConstraint {
			continue
		}

		constraint, ok := node.AlterTableCmd.Def.Node.(*pgQuery.Node_Constraint)
		if !ok {
			continue
		}

		// Handle primary key constraints
		if constraint.Constraint.Contype == pgQuery.ConstrType_CONSTR_PRIMARY {
			// Get the primary key column names from Keys
			for _, key := range constraint.Constraint.Keys {
				node, ok := key.Node.(*pgQuery.Node_String_)
				if !ok {
					continue
				}

				// Find the column and add primary constraint
				for _, col := range sourceTable.Columns {
					if col.Name == node.String_.Sval {
						col.Constraints = append(col.Constraints, "primary")
						break
					}
				}
			}
			continue
		}

		if constraint.Constraint.Contype != pgQuery.ConstrType_CONSTR_FOREIGN {
			continue
		}

		foreignReference := &ForeignReference{
			Table: constraint.Constraint.Pktable.Relname,
		}

		// Get the referenced column name from PkAttrs (primary key attributes)
		for _, pkattr := range constraint.Constraint.PkAttrs {
			node, ok := pkattr.Node.(*pgQuery.Node_String_)
			if !ok {
				continue
			}
			foreignReference.Column = node.String_.Sval
		}

		// Get the foreign key column name from FkAttrs (foreign key attributes)
		for _, fkattr := range constraint.Constraint.FkAttrs {
			node, ok := fkattr.Node.(*pgQuery.Node_String_)
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

			// Check if column was found before trying to append to it
			if column != nil {
				column.ForeignKeyReferences = append(column.ForeignKeyReferences, foreignReference)
			}
		}
	}

	return nil
}

func toTable(stmt *pgQuery.Node_CreateStmt) *Table {
	table := &Table{
		Name: stmt.CreateStmt.Relation.Relname,
	}

	for _, columnNode := range stmt.CreateStmt.TableElts {
		columnNodeDefinition, ok := columnNode.Node.(*pgQuery.Node_ColumnDef)
		if !ok {
			continue
		}

		table.Columns = append(table.Columns, generateColumnProperties(columnNodeDefinition.ColumnDef))
	}

	return table
}

func generateColumnProperties(columnDefinition *pgQuery.ColumnDef) *Column {
	column := &Column{
		Name: columnDefinition.Colname,
	}

	for _, node := range columnDefinition.TypeName.Names {
		stringNode, ok := node.Node.(*pgQuery.Node_String_)
		if !ok {
			fmt.Printf("unknown name node %v\n", stringNode)
			continue
		}

		column.Type = stringNode.String_.Sval

		// get length value
		for _, mod := range columnDefinition.TypeName.Typmods {
			aConst, ok := mod.Node.(*pgQuery.Node_AConst)
			if !ok {
				continue
			}

			integer, ok := aConst.AConst.Val.(*pgQuery.A_Const_Ival)
			if !ok {
				continue
			}

			column.Length = int(integer.Ival.GetIval())
		}

		// map not null constraint to string
		for _, constraint := range columnDefinition.Constraints {
			nodeConstraint, ok := constraint.Node.(*pgQuery.Node_Constraint)
			if !ok {
				continue
			}

			switch nodeConstraint.Constraint.Contype {
			case pgQuery.ConstrType_CONSTR_PRIMARY:
				column.Constraints = append(column.Constraints, "primary")
			case pgQuery.ConstrType_CONSTR_FOREIGN:
				foreignReference := &ForeignReference{
					Table: nodeConstraint.Constraint.Pktable.Relname,
				}

				var found bool
				for _, pkattr := range nodeConstraint.Constraint.PkAttrs {
					node, ok := pkattr.Node.(*pgQuery.Node_String_)
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
			case pgQuery.ConstrType_CONSTR_NOTNULL:
				var alreadyExists bool
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

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
