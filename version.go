// Package semver implements parsing, comparison, and ordering of
// Semantic Versioning 2.0.0 version strings (see semver.org).
package semver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Version holds the parsed parts of a semantic version string.
// Prerelease and Build are the raw dot-separated identifier lists,
// without the leading "-" or "+".
type Version struct {
	Major, Minor, Patch uint64
	Prerelease          string
	Build               string
}

// ErrInvalidVersion is returned by Parse when the input does not
// conform to the semver 2.0.0 grammar.
var ErrInvalidVersion = errors.New("semver: invalid version string")

// Parse converts a version string into a Version. A single leading
// "v" is tolerated (e.g. "v1.2.3") since it's common in tags, but is
// not part of the semver spec itself.
func Parse(s string) (Version, error) {
	rest := strings.TrimPrefix(s, "v")

	var v Version

	if i := strings.IndexByte(rest, '+'); i >= 0 {
		v.Build = rest[i+1:]
		rest = rest[:i]
		if !validDotted(v.Build, isBuildIdentifier) {
			return Version{}, ErrInvalidVersion
		}
	}
	if i := strings.IndexByte(rest, '-'); i >= 0 {
		v.Prerelease = rest[i+1:]
		rest = rest[:i]
		if !validDotted(v.Prerelease, isPrereleaseIdentifier) {
			return Version{}, ErrInvalidVersion
		}
	}

	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return Version{}, ErrInvalidVersion
	}
	nums := [3]uint64{}
	for i, p := range parts {
		if !isNumericIdentifier(p) {
			return Version{}, ErrInvalidVersion
		}
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return Version{}, ErrInvalidVersion
		}
		nums[i] = n
	}
	v.Major, v.Minor, v.Patch = nums[0], nums[1], nums[2]
	return v, nil
}

// String renders the version back into its canonical textual form.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare returns -1, 0, or 1 depending on whether v is less than,
// equal to, or greater than o, following semver precedence rules.
// Build metadata is ignored, as required by the spec.
func (v Version) Compare(o Version) int {
	if v.Major != o.Major {
		return cmpUint(v.Major, o.Major)
	}
	if v.Minor != o.Minor {
		return cmpUint(v.Minor, o.Minor)
	}
	if v.Patch != o.Patch {
		return cmpUint(v.Patch, o.Patch)
	}
	return comparePrerelease(v.Prerelease, o.Prerelease)
}

// Less reports whether a sorts before b.
func Less(a, b Version) bool {
	return a.Compare(b) < 0
}

func cmpUint(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// comparePrerelease implements the semver rule that a version without
// a prerelease has higher precedence than one with, and otherwise
// compares identifiers left to right.
func comparePrerelease(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	aIDs := strings.Split(a, ".")
	bIDs := strings.Split(b, ".")
	for i := 0; i < len(aIDs) && i < len(bIDs); i++ {
		if c := compareIdentifier(aIDs[i], bIDs[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(aIDs), len(bIDs))
}

// compareIdentifier compares a single dot-separated prerelease field.
// Numeric identifiers compare numerically and always sort below
// alphanumeric ones; alphanumeric identifiers compare lexically.
func compareIdentifier(a, b string) int {
	aNum, bNum := isNumericOnly(a), isNumericOnly(b)
	switch {
	case aNum && bNum:
		an, _ := strconv.ParseUint(a, 10, 64)
		bn, _ := strconv.ParseUint(b, 10, 64)
		return cmpUint(an, bn)
	case aNum:
		return -1
	case bNum:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func validDotted(s string, valid func(string) bool) bool {
	if s == "" {
		return false
	}
	for _, id := range strings.Split(s, ".") {
		if !valid(id) {
			return false
		}
	}
	return true
}

// isNumericIdentifier reports whether s is a valid numeric core
// component: digits only, and no leading zero unless the value is 0.
func isNumericIdentifier(s string) bool {
	if !isNumericOnly(s) {
		return false
	}
	return len(s) == 1 || s[0] != '0'
}

func isNumericOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isPrereleaseIdentifier allows alphanumeric identifiers, and numeric
// identifiers without leading zeros.
func isPrereleaseIdentifier(s string) bool {
	if isNumericOnly(s) {
		return isNumericIdentifier(s)
	}
	return isAlphanumericIdentifier(s)
}

// isBuildIdentifier allows any alphanumeric identifier, leading
// zeros included, per the spec.
func isBuildIdentifier(s string) bool {
	return isAlphanumericIdentifier(s)
}

func isAlphanumericIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		ok := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-'
		if !ok {
			return false
		}
	}
	return true
}
