package semver

import "testing"

func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", s, err)
	}
	return v
}

func TestRangeMatches(t *testing.T) {
	cases := []struct {
		rng  string
		in   string
		want bool
	}{
		{">=1.0.0 <2.0.0", "1.5.0", true},
		{">=1.0.0 <2.0.0", "2.0.0", false},
		{">=1.0.0 <2.0.0", "0.9.9", false},
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},
		{"=1.2.3", "1.2.3", true},
		{"!=1.2.3", "1.2.3", false},
		{"!=1.2.3", "1.2.4", true},
		{">1.0.0", "1.0.1", true},
		{">1.0.0", "1.0.0", false},
		{"<=1.0.0", "1.0.0", true},
		{"^1.2.3", "1.9.9", true},
		{"^1.2.3", "2.0.0", false},
		{"^1.2.3", "1.2.2", false},
		{"^0.2.3", "0.2.9", true},
		{"^0.2.3", "0.3.0", false},
		{"^0.0.3", "0.0.3", true},
		{"^0.0.3", "0.0.4", false},
		{"~1.2.3", "1.2.9", true},
		{"~1.2.3", "1.3.0", false},
		{"~1.2.3", "1.2.2", false},
		{">=1.0.0 <2.0.0 || >=3.0.0", "3.5.0", true},
		{">=1.0.0 <2.0.0 || >=3.0.0", "2.5.0", false},
	}
	for _, c := range cases {
		r, err := ParseRange(c.rng)
		if err != nil {
			t.Fatalf("ParseRange(%q) returned error: %v", c.rng, err)
		}
		v := mustParse(t, c.in)
		if got := r.Matches(v); got != c.want {
			t.Errorf("ParseRange(%q).Matches(%q) = %v, want %v", c.rng, c.in, got, c.want)
		}
	}
}

func TestParseRangeInvalid(t *testing.T) {
	cases := []string{
		"",
		"   ",
		">=1.0.0 <2.0.0 ||",
		">=abc",
		"^abc",
		"~1.2",
		">=",
	}
	for _, in := range cases {
		if _, err := ParseRange(in); err == nil {
			t.Errorf("ParseRange(%q) succeeded, want error", in)
		}
	}
}
