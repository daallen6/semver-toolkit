package semver

import (
	"errors"
	"testing"
)

func TestParseValid(t *testing.T) {
	cases := []struct {
		in   string
		want Version
	}{
		{"0.0.0", Version{0, 0, 0, "", ""}},
		{"1.2.3", Version{1, 2, 3, "", ""}},
		{"v1.2.3", Version{1, 2, 3, "", ""}},
		{"10.20.30", Version{10, 20, 30, "", ""}},
		{"1.2.3-alpha", Version{1, 2, 3, "alpha", ""}},
		{"1.2.3-alpha.1", Version{1, 2, 3, "alpha.1", ""}},
		{"1.2.3-0.3.7", Version{1, 2, 3, "0.3.7", ""}},
		{"1.2.3-x-y-z.--", Version{1, 2, 3, "x-y-z.--", ""}},
		{"1.2.3+build.5", Version{1, 2, 3, "", "build.5"}},
		{"1.2.3+0001", Version{1, 2, 3, "", "0001"}},
		{"1.2.3-beta+build.2", Version{1, 2, 3, "beta", "build.2"}},
		{"1.2.3-rc.1+build.123", Version{1, 2, 3, "rc.1", "build.123"}},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []string{
		"",
		"1",
		"1.2",
		"1.2.3.4",
		"1.2.3-",
		"1.2.3+",
		"01.2.3",
		"1.02.3",
		"1.2.03",
		"-1.2.3",
		"1.-2.3",
		"1.2.-3",
		"1.2.3-01",
		"1.2.3-alpha..1",
		"1.2.3-alpha_beta",
		"1.2.3+_",
		"a.b.c",
		"1.2.3+build_meta",
	}
	for _, in := range cases {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", in)
		}
	}
}

func TestStringRoundTrip(t *testing.T) {
	cases := []string{
		"0.0.0",
		"1.2.3",
		"1.2.3-alpha.1",
		"1.2.3+build.5",
		"1.2.3-rc.1+build.123",
	}
	for _, in := range cases {
		v, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", in, err)
		}
		if got := v.String(); got != in {
			t.Errorf("Parse(%q).String() = %q, want %q", in, got, in)
		}
	}
}

func TestStringDropsLeadingV(t *testing.T) {
	v, err := Parse("v1.2.3")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got, want := v.String(), "1.2.3"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// TestPrereleaseOrdering follows the example precedence chain from
// the semver 2.0.0 spec (section 11).
func TestPrereleaseOrdering(t *testing.T) {
	order := []string{
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
		"1.0.0",
	}
	for i := 0; i < len(order)-1; i++ {
		a, err := Parse(order[i])
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", order[i], err)
		}
		b, err := Parse(order[i+1])
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", order[i+1], err)
		}
		if !Less(a, b) {
			t.Errorf("expected %q < %q", order[i], order[i+1])
		}
		if a.Compare(b) != -1 {
			t.Errorf("Compare(%q, %q) = %d, want -1", order[i], order[i+1], a.Compare(b))
		}
		if b.Compare(a) != 1 {
			t.Errorf("Compare(%q, %q) = %d, want 1", order[i+1], order[i], b.Compare(a))
		}
	}
}

func TestCompareIgnoresBuild(t *testing.T) {
	a, err := Parse("1.2.3+build.1")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	b, err := Parse("1.2.3+build.2")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if c := a.Compare(b); c != 0 {
		t.Errorf("Compare with differing build metadata = %d, want 0", c)
	}
}

func TestBump(t *testing.T) {
	cases := []struct {
		in   string
		kind string
		want string
	}{
		{"1.2.3", "major", "2.0.0"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3-alpha.1", "major", "2.0.0"},
		{"1.2.3-alpha.1", "minor", "1.3.0"},
		{"1.2.3-alpha.1", "patch", "1.2.4"},
		{"1.2.3+build.5", "patch", "1.2.4"},
		{"1.2.3", "prerelease", "1.2.4-0"},
		{"1.2.3-alpha", "prerelease", "1.2.3-alpha.0"},
		{"1.2.3-alpha.1", "prerelease", "1.2.3-alpha.2"},
		{"1.2.3-alpha.9", "prerelease", "1.2.3-alpha.10"},
		{"1.2.3-0.9", "prerelease", "1.2.3-0.10"},
	}
	for _, c := range cases {
		v, err := Parse(c.in)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", c.in, err)
		}
		got, err := v.Bump(c.kind)
		if err != nil {
			t.Fatalf("Bump(%q).Bump(%q) returned error: %v", c.in, c.kind, err)
		}
		if got.String() != c.want {
			t.Errorf("%q.Bump(%q) = %q, want %q", c.in, c.kind, got.String(), c.want)
		}
	}
}

func TestBumpInvalidKind(t *testing.T) {
	v, err := Parse("1.2.3")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if _, err := v.Bump("bogus"); !errors.Is(err, ErrInvalidBumpKind) {
		t.Errorf("Bump(%q) error = %v, want ErrInvalidBumpKind", "bogus", err)
	}
}

func TestBumpIsMonotonic(t *testing.T) {
	kinds := []string{"major", "minor", "patch", "prerelease"}
	starts := []string{"0.0.0", "1.2.3", "1.2.3-alpha", "1.2.3-alpha.1", "1.2.3-0.9"}
	for _, start := range starts {
		v, err := Parse(start)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", start, err)
		}
		for _, kind := range kinds {
			next, err := v.Bump(kind)
			if err != nil {
				t.Fatalf("%q.Bump(%q) returned error: %v", start, kind, err)
			}
			if !Less(v, next) {
				t.Errorf("%q.Bump(%q) = %q, want a version greater than %q", start, kind, next.String(), start)
			}
		}
	}
}

func TestCompareCoreParts(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.0", 0},
		{"1.1.0", "1.2.0", -1},
		{"1.2.1", "1.2.0", 1},
		{"1.9.0", "1.10.0", -1},
	}
	for _, c := range cases {
		a, err := Parse(c.a)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", c.a, err)
		}
		b, err := Parse(c.b)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", c.b, err)
		}
		if got := a.Compare(b); got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
