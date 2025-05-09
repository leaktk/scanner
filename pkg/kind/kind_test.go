package kind

import (
	"testing"
)

// TestKindsMatch tests the KindsMatch function
func TestKindsMatch(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "exact match, lowercase",
			a:    "mykind",
			b:    "mykind",
			want: true,
		},
		{
			name: "exact match, mixed case (should normalize to same)",
			a:    "MyKind",
			b:    "mykind",
			want: true,
		},
		{
			name: "match with dashes",
			a:    "my-kind",
			b:    "mykind",
			want: true,
		},
		{
			name: "match with underscores",
			a:    "my_kind",
			b:    "mykind",
			want: true,
		},
		{
			name: "match with mixed dashes, underscores, and case",
			a:    "My-Complex_Kind",
			b:    "mycomplexkind",
			want: true,
		},
		{
			name: "match with different normalization needed for both",
			a:    "My-First_Kind",
			b:    "myfirst-kind",
			want: true,
		},
		{
			name: "no match",
			a:    "mykind",
			b:    "anotherkind",
			want: false,
		},
		{
			name: "no match, even with normalization",
			a:    "My-Kind",
			b:    "Another_Type",
			want: false,
		},
		{
			name: "empty strings",
			a:    "",
			b:    "",
			want: true,
		},
		{
			name: "one empty, one not",
			a:    "",
			b:    "mykind",
			want: false,
		},
		{
			name: "one empty, one not (reversed)",
			a:    "mykind",
			b:    "",
			want: false,
		},
		{
			name: "different strings that normalize to same non-empty string",
			a:    "Test-Resource",
			b:    "test_resource",
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := KindsMatch(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("KindsMatch(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
