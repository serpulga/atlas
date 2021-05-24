package mysql

import (
	"testing"

	"ariga.io/atlas/sql/schema"

	"github.com/stretchr/testify/require"
)

func TestDiff_TableDiff(t *testing.T) {
	type testcase struct {
		name        string
		from, to    *schema.Table
		wantChanges []schema.Change
		wantErr     bool
	}
	tests := []testcase{
		{
			name: "no changes",
			from: &schema.Table{Name: "users"},
			to:   &schema.Table{Name: "users"},
		},
		{
			name: "change primary key",
			from: func() *schema.Table {
				t := &schema.Table{Name: "users", Columns: []*schema.Column{{Name: "id", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}}}}
				t.PrimaryKey = t.Columns
				return t
			}(),
			to:      &schema.Table{Name: "users"},
			wantErr: true,
		},
		{
			name: "add collation",
			from: &schema.Table{Name: "users"},
			to:   &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Collation{V: "latin1"}}},
			wantChanges: []schema.Change{
				&schema.AddAttr{
					A: &schema.Collation{V: "latin1"},
				},
			},
		},
		{
			name: "drop collation",
			from: &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Collation{V: "latin1"}}},
			to:   &schema.Table{Name: "users"},
			wantChanges: []schema.Change{
				&schema.DropAttr{
					A: &schema.Collation{V: "latin1"},
				},
			},
		},
		{
			name: "modify collation",
			from: &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Collation{V: "utf8"}}},
			to:   &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Collation{V: "latin1"}}},
			wantChanges: []schema.Change{
				&schema.ModifyAttr{
					From: &schema.Collation{V: "utf8"},
					To:   &schema.Collation{V: "latin1"},
				},
			},
		},
		{
			name: "add charset",
			from: &schema.Table{Name: "users"},
			to:   &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Charset{V: "hebrew"}}},
			wantChanges: []schema.Change{
				&schema.AddAttr{
					A: &schema.Charset{V: "hebrew"},
				},
			},
		},
		{
			name: "drop charset",
			from: &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Charset{V: "hebrew"}}},
			to:   &schema.Table{Name: "users"},
			wantChanges: []schema.Change{
				&schema.DropAttr{
					A: &schema.Charset{V: "hebrew"},
				},
			},
		},
		{
			name: "modify charset",
			from: &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Charset{V: "hebrew"}}},
			to:   &schema.Table{Name: "users", Attrs: []schema.Attr{&schema.Charset{V: "binary"}}},
			wantChanges: []schema.Change{
				&schema.ModifyAttr{
					From: &schema.Charset{V: "hebrew"},
					To:   &schema.Charset{V: "binary"},
				},
			},
		},
		{
			name: "add check",
			from: &schema.Table{Name: "t1"},
			to:   &schema.Table{Name: "t1", Attrs: []schema.Attr{&Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')"}}},
			wantChanges: []schema.Change{
				&schema.AddAttr{
					A: &Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')"},
				},
			},
		},
		{
			name: "drop check",
			from: &schema.Table{Name: "t1", Attrs: []schema.Attr{&Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')"}}},
			to:   &schema.Table{Name: "t1"},
			wantChanges: []schema.Change{
				&schema.DropAttr{
					A: &Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')"},
				},
			},
		},
		{
			name: "modify check",
			from: &schema.Table{Name: "t1", Attrs: []schema.Attr{&Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')"}}},
			to:   &schema.Table{Name: "t1", Attrs: []schema.Attr{&Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')", Enforced: true}}},
			wantChanges: []schema.Change{
				&schema.ModifyAttr{
					From: &Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')"},
					To:   &Check{Name: "users_chk1_c1", Clause: "(`c1` <>_latin1\\'foo\\')", Enforced: true},
				},
			},
		},
		func() testcase {
			var (
				from = &schema.Table{
					Name: "t1",
					Columns: []*schema.Column{
						{Name: "c1", Type: &schema.ColumnType{Raw: "json", Type: &schema.JSONType{T: "json"}}},
						{Name: "c2", Type: &schema.ColumnType{Raw: "tinyint", Type: &schema.IntegerType{T: "tinyint"}}},
					},
				}
				to = &schema.Table{
					Name: "t1",
					Columns: []*schema.Column{
						{
							Name:    "c1",
							Type:    &schema.ColumnType{Raw: "json", Type: &schema.JSONType{T: "json"}, Null: true},
							Default: &schema.RawExpr{X: "{}"},
							Attrs:   []schema.Attr{&schema.Comment{Text: "json comment"}},
						},
						{Name: "c3", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				}
			)
			return testcase{
				name: "columns",
				from: from,
				to:   to,
				wantChanges: []schema.Change{
					&schema.ModifyColumn{
						From:   from.Columns[0],
						To:     to.Columns[0],
						Change: schema.ChangeNull | schema.ChangeComment | schema.ChangeDefault,
					},
					&schema.DropColumn{C: from.Columns[1]},
					&schema.AddColumn{C: to.Columns[1]},
				},
			}
		}(),
		func() testcase {
			var (
				from = &schema.Table{
					Name: "t1",
					Columns: []*schema.Column{
						{Name: "c1", Type: &schema.ColumnType{Raw: "json", Type: &schema.JSONType{T: "json"}}},
						{Name: "c2", Type: &schema.ColumnType{Raw: "tinyint", Type: &schema.IntegerType{T: "tinyint"}}},
						{Name: "c3", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				}
				to = &schema.Table{
					Name: "t1",
					Columns: []*schema.Column{
						{Name: "c1", Type: &schema.ColumnType{Raw: "json", Type: &schema.JSONType{T: "json"}}},
						{Name: "c2", Type: &schema.ColumnType{Raw: "tinyint", Type: &schema.IntegerType{T: "tinyint"}}},
						{Name: "c3", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				}
			)
			from.Indexes = []*schema.Index{
				{Name: "c1_index", Unique: true, Table: from, Parts: []*schema.IndexPart{{SeqNo: 1, C: from.Columns[0]}}},
				{Name: "c2_unique", Unique: true, Table: from, Parts: []*schema.IndexPart{{SeqNo: 1, C: from.Columns[1]}}},
			}
			to.Indexes = []*schema.Index{
				{Name: "c1_index", Table: from, Parts: []*schema.IndexPart{{SeqNo: 1, C: from.Columns[0]}}},
				{Name: "c3_unique", Unique: true, Table: from, Parts: []*schema.IndexPart{{SeqNo: 1, C: to.Columns[1]}}},
			}
			return testcase{
				name: "indexes",
				from: from,
				to:   to,
				wantChanges: []schema.Change{
					&schema.ModifyIndex{From: from.Indexes[0], To: to.Indexes[0], Change: schema.ChangeUnique},
					&schema.DropIndex{I: from.Indexes[1]},
					&schema.AddIndex{I: to.Indexes[1]},
				},
			}
		}(),
		func() testcase {
			var (
				ref = &schema.Table{
					Name: "t2",
					Columns: []*schema.Column{
						{Name: "id", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
						{Name: "ref_id", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				}
				from = &schema.Table{
					Name: "t1",
					Columns: []*schema.Column{
						{Name: "t2_id", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				}
				to = &schema.Table{
					Name: "t1",
					Columns: []*schema.Column{
						{Name: "t2_id", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				}
			)
			from.ForeignKeys = []*schema.ForeignKey{
				{Table: from, Columns: from.Columns, RefTable: ref, RefColumns: ref.Columns[:1]},
			}
			to.ForeignKeys = []*schema.ForeignKey{
				{Table: to, Columns: to.Columns, RefTable: ref, RefColumns: ref.Columns[1:]},
			}
			return testcase{
				name: "indexes",
				from: from,
				to:   to,
				wantChanges: []schema.Change{
					&schema.ModifyForeignKey{
						From:   from.ForeignKeys[0],
						To:     to.ForeignKeys[0],
						Change: schema.ChangeRefColumn,
					},
				},
			}
		}(),
	}
	for _, tt := range tests {
		var d Diff
		t.Run(tt.name, func(t *testing.T) {
			changes, err := d.TableDiff(tt.from, tt.to)
			require.Equal(t, tt.wantErr, err != nil)
			require.EqualValues(t, tt.wantChanges, changes)
		})
	}
}

func TestDiff_SchemaDiff(t *testing.T) {
	var (
		d    Diff
		from = &schema.Schema{
			Tables: []*schema.Table{
				{Name: "users"},
				{Name: "pets"},
			},
			Attrs: []schema.Attr{
				&schema.Collation{V: "latin1"},
			},
		}
		to = &schema.Schema{
			Tables: []*schema.Table{
				{
					Name: "users",
					Columns: []*schema.Column{
						{Name: "t2_id", Type: &schema.ColumnType{Raw: "int", Type: &schema.IntegerType{T: "int"}}},
					},
				},
				{Name: "groups"},
			},
			Attrs: []schema.Attr{
				&schema.Collation{V: "utf8"},
			},
		}
	)
	changes, err := d.SchemaDiff(from, to)
	require.NoError(t, err)
	require.EqualValues(t, []schema.Change{
		&schema.ModifyAttr{From: from.Attrs[0], To: to.Attrs[0]},
		&schema.ModifyTable{T: from.Tables[0], Changes: []schema.Change{&schema.AddColumn{C: to.Tables[0].Columns[0]}}},
		&schema.DropTable{T: from.Tables[1]},
		&schema.AddTable{T: to.Tables[1]},
	}, changes)
}