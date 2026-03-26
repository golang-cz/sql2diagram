package schema

import (
	"fmt"
	"slices"

	pgQuery "github.com/pganalyze/pg_query_go/v6"
)

type Schema struct {
	Tables []*Table
}

type ForeignReference struct {
	Table  string
	Column string
}

type Table struct {
	Name              string
	Columns           []*Column
	UniqueConstraints [][]string
}

type Constraints []string

func (c *Constraints) Has(value string) bool  { return slices.Contains(*c, value) }
func (c *Constraints) Add(value string)        { if !c.Has(value) { *c = append(*c, value) } }

type Column struct {
	Name                 string
	Type                 string
	Constraints          Constraints
	ForeignKeyReferences []*ForeignReference
	Length               int
}

func Parse(input string) (*Schema, error) {
	tree, err := pgQuery.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("parse SQL: %w", err)
	}
	return astTreeToSchema(tree)
}

func astTreeToSchema(tree *pgQuery.ParseResult) (*Schema, error) {
	schema := &Schema{}

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
	sourceTable := schema.findTable(stmt.Relation.Relname)
	if sourceTable == nil {
		return fmt.Errorf("table %q not found in schema", stmt.Relation.Relname)
	}

	for _, cmd := range stmt.Cmds {
		node, ok := cmd.Node.(*pgQuery.Node_AlterTableCmd)
		if !ok || node.AlterTableCmd.Subtype != pgQuery.AlterTableType_AT_AddConstraint {
			continue
		}

		constraint, ok := node.AlterTableCmd.Def.Node.(*pgQuery.Node_Constraint)
		if !ok {
			continue
		}

		cols := constraintColumns(constraint.Constraint.Keys)

		switch constraint.Constraint.Contype {
		case pgQuery.ConstrType_CONSTR_PRIMARY:
			for _, name := range cols {
				if col := sourceTable.findColumn(name); col != nil {
					col.Constraints.Add("primary")
				}
			}
		case pgQuery.ConstrType_CONSTR_UNIQUE:
			sourceTable.addUniqueConstraint(cols)
		case pgQuery.ConstrType_CONSTR_FOREIGN:
			fk := &ForeignReference{Table: constraint.Constraint.Pktable.Relname}
			for _, pkattr := range constraint.Constraint.PkAttrs {
				if n, ok := pkattr.Node.(*pgQuery.Node_String_); ok {
					fk.Column = n.String_.Sval
				}
			}
			for _, name := range constraintColumns(constraint.Constraint.FkAttrs) {
				if col := sourceTable.findColumn(name); col != nil {
					col.ForeignKeyReferences = append(col.ForeignKeyReferences, fk)
				}
			}
		}
	}

	return nil
}

func toTable(stmt *pgQuery.Node_CreateStmt) *Table {
	table := &Table{Name: stmt.CreateStmt.Relation.Relname}
	var pendingUniques [][]string

	for _, columnNode := range stmt.CreateStmt.TableElts {
		if colDef, ok := columnNode.Node.(*pgQuery.Node_ColumnDef); ok {
			table.Columns = append(table.Columns, generateColumnProperties(colDef.ColumnDef))
			continue
		}

		constraint, ok := columnNode.Node.(*pgQuery.Node_Constraint)
		if !ok || constraint.Constraint.Contype != pgQuery.ConstrType_CONSTR_UNIQUE {
			continue
		}
		pendingUniques = append(pendingUniques, constraintColumns(constraint.Constraint.Keys))
	}

	for _, cols := range pendingUniques {
		table.addUniqueConstraint(cols)
	}

	return table
}

func generateColumnProperties(columnDefinition *pgQuery.ColumnDef) *Column {
	column := &Column{Name: columnDefinition.Colname}

	for _, node := range columnDefinition.TypeName.Names {
		stringNode, ok := node.Node.(*pgQuery.Node_String_)
		if !ok {
			continue
		}

		column.Type = stringNode.String_.Sval

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

		for _, constraint := range columnDefinition.Constraints {
			nodeConstraint, ok := constraint.Node.(*pgQuery.Node_Constraint)
			if !ok {
				continue
			}

			switch nodeConstraint.Constraint.Contype {
			case pgQuery.ConstrType_CONSTR_PRIMARY:
				column.Constraints.Add("primary")
			case pgQuery.ConstrType_CONSTR_UNIQUE:
				column.Constraints.Add("unique")
			case pgQuery.ConstrType_CONSTR_NOTNULL:
				column.Constraints.Add("not null")
			case pgQuery.ConstrType_CONSTR_FOREIGN:
				column.addForeignKey(nodeConstraint.Constraint)
			}
		}
	}

	return column
}

func (s *Schema) findTable(name string) *Table {
	for _, t := range s.Tables {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (t *Table) findColumn(name string) *Column {
	for _, col := range t.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

func (t *Table) addUniqueConstraint(columns []string) {
	if len(columns) == 1 {
		if col := t.findColumn(columns[0]); col != nil {
			col.Constraints.Add("unique")
		}
	} else if len(columns) > 1 && !slices.ContainsFunc(t.UniqueConstraints, func(e []string) bool { return slices.Equal(e, columns) }) {
		t.UniqueConstraints = append(t.UniqueConstraints, columns)
	}
}

func (c *Column) addForeignKey(con *pgQuery.Constraint) {
	fk := &ForeignReference{Table: con.Pktable.Relname}
	for _, pkattr := range con.PkAttrs {
		if n, ok := pkattr.Node.(*pgQuery.Node_String_); ok {
			fk.Column = n.String_.Sval
		}
	}
	if !slices.ContainsFunc(c.ForeignKeyReferences, func(r *ForeignReference) bool {
		return r.Table == fk.Table && r.Column == fk.Column
	}) {
		c.ForeignKeyReferences = append(c.ForeignKeyReferences, fk)
	}
}

func constraintColumns(keys []*pgQuery.Node) []string {
	cols := make([]string, 0, len(keys))
	for _, key := range keys {
		if n, ok := key.Node.(*pgQuery.Node_String_); ok {
			cols = append(cols, n.String_.Sval)
		}
	}
	return cols
}
