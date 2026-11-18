package db

import "testing"

func TestMigrateURL(t *testing.T) {
	tests := []struct{ in, want string }{
		{"postgres://u:p@h:5432/d", "pgx5://u:p@h:5432/d"},
		{"postgresql://u:p@h:5432/d?sslmode=disable", "pgx5://u:p@h:5432/d?sslmode=disable"},
		{"pgx5://u:p@h/d", "pgx5://u:p@h/d"},
	}
	for _, tc := range tests {
		if got := migrateURL(tc.in); got != tc.want {
			t.Errorf("migrateURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
