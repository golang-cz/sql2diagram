package schema

import "testing"

func TestParseTracksUniqueConstraints(t *testing.T) {
	t.Parallel()

	schema, err := Parse(`
CREATE TABLE demo (
    email text UNIQUE,
    org_id text,
    name text,
    UNIQUE (org_id, name)
);
`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(schema.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(schema.Tables))
	}

	table := schema.Tables[0]
	if !table.Columns[0].Constraints.Has("unique") {
		t.Fatalf("expected email column to be marked unique, got %v", table.Columns[0].Constraints)
	}

	if len(table.UniqueConstraints) != 1 {
		t.Fatalf("expected 1 composite unique constraint, got %d", len(table.UniqueConstraints))
	}
}

func TestParseAppliesSingleColumnTableUniqueDeclaredBeforeColumn(t *testing.T) {
	t.Parallel()

	schema, err := Parse(`
CREATE TABLE demo (
    CONSTRAINT demo_email_key UNIQUE (email),
    email text,
    name text
);
`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(schema.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(schema.Tables))
	}

	email := schema.Tables[0].Columns[0]
	if email.Name != "email" {
		t.Fatalf("expected first column to be email, got %q", email.Name)
	}

	if !email.Constraints.Has("unique") {
		t.Fatalf("expected email column to include unique constraint, got %v", email.Constraints)
	}
}

func TestParseAlterTableForeignKeyAssignsReferenceToSourceColumn(t *testing.T) {
	t.Parallel()

	schema, err := Parse(`
CREATE TABLE parent (
    id uuid PRIMARY KEY
);

CREATE TABLE child (
    parent_id uuid
);

ALTER TABLE child
    ADD CONSTRAINT child_parent_id_fkey
    FOREIGN KEY (parent_id) REFERENCES parent (id);
`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(schema.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(schema.Tables))
	}

	child := schema.findTable("child")
	if child == nil {
		t.Fatal("expected child table to exist")
	}

	parentID := child.findColumn("parent_id")
	if parentID == nil {
		t.Fatal("expected child.parent_id column to exist")
	}

	if len(parentID.ForeignKeyReferences) != 1 {
		t.Fatalf("expected 1 foreign key reference on child.parent_id, got %d", len(parentID.ForeignKeyReferences))
	}

	ref := parentID.ForeignKeyReferences[0]
	if ref.Table != "parent" || ref.Column != "id" {
		t.Fatalf("expected child.parent_id to reference parent.id, got %s.%s", ref.Table, ref.Column)
	}
}
