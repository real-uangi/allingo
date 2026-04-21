package db

import (
	"testing"

	"gorm.io/gorm/clause"
)

func TestParseOrderByColumns(t *testing.T) {
	allowed := map[string]struct{}{
		"id":         {},
		"created_at": {},
		"updated_at": {},
	}

	tests := []struct {
		name string
		raw  string
		want []clause.OrderByColumn
	}{
		{
			name: "empty use fallback",
			raw:  "",
			want: []clause.OrderByColumn{
				{Column: clause.Column{Name: "id"}},
			},
		},
		{
			name: "single field asc by default",
			raw:  "created_at",
			want: []clause.OrderByColumn{
				{Column: clause.Column{Name: "created_at"}},
			},
		},
		{
			name: "multi field with explicit direction",
			raw:  "id desc, created_at asc",
			want: []clause.OrderByColumn{
				{Column: clause.Column{Name: "id"}, Desc: true},
				{Column: clause.Column{Name: "created_at"}},
			},
		},
		{
			name: "allow omitted direction",
			raw:  "updated_at, id desc",
			want: []clause.OrderByColumn{
				{Column: clause.Column{Name: "updated_at"}},
				{Column: clause.Column{Name: "id"}, Desc: true},
			},
		},
		{
			name: "ignore invalid and keep valid",
			raw:  "id desc, not_exists asc, created_at",
			want: []clause.OrderByColumn{
				{Column: clause.Column{Name: "id"}, Desc: true},
				{Column: clause.Column{Name: "created_at"}},
			},
		},
		{
			name: "all invalid fallback",
			raw:  "id;drop table x, unknown xx",
			want: []clause.OrderByColumn{
				{Column: clause.Column{Name: "id"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOrderByColumns(tt.raw, allowed, "id")
			if len(got) != len(tt.want) {
				t.Fatalf("parseOrderByColumns(%q) len=%d, want %d", tt.raw, len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Column.Name != tt.want[i].Column.Name || got[i].Desc != tt.want[i].Desc {
					t.Fatalf("parseOrderByColumns(%q)[%d] = (%s,%v), want (%s,%v)", tt.raw, i, got[i].Column.Name, got[i].Desc, tt.want[i].Column.Name, tt.want[i].Desc)
				}
			}
		})
	}
}
