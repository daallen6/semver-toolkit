package semver

import (
	"errors"
	"strings"
)

// Range is a version constraint of the kind commonly written in
// dependency manifests: a set of comparators such as ">=1.0.0
// <2.0.0", optionally with several such sets joined by "||" so a
// version needs to satisfy only one of them.
//
// Caret (^1.2.3) and tilde (~1.2.3) are shorthand for a lower and
// upper bound pair, expanded into comparators at parse time using
// the convention popularized by npm's node-semver: caret allows
// changes that don't touch the leftmost nonzero component, tilde
// allows patch-level changes only.
type Range struct {
	sets [][]rangeComparator
}

// ErrInvalidRange is returned by ParseRange when the input isn't a
// well-formed comparator set.
var ErrInvalidRange = errors.New("semver: invalid range")

type rangeOp int

const (
	opEQ rangeOp = iota
	opNE
	opLT
	opLE
	opGT
	opGE
)

type rangeComparator struct {
	op rangeOp
	v  Version
}

// ParseRange parses a range expression into a Range that can be
// tested against a Version with Range.Matches.
func ParseRange(s string) (Range, error) {
	var r Range
	for _, part := range strings.Split(s, "||") {
		set, err := parseComparatorSet(part)
		if err != nil {
			return Range{}, err
		}
		r.sets = append(r.sets, set)
	}
	return r, nil
}

func parseComparatorSet(s string) ([]rangeComparator, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, ErrInvalidRange
	}
	var set []rangeComparator
	for _, f := range fields {
		cs, err := parseComparatorToken(f)
		if err != nil {
			return nil, err
		}
		set = append(set, cs...)
	}
	return set, nil
}

func parseComparatorToken(tok string) ([]rangeComparator, error) {
	switch {
	case strings.HasPrefix(tok, "^"):
		v, err := Parse(tok[1:])
		if err != nil {
			return nil, ErrInvalidRange
		}
		return caretComparators(v), nil
	case strings.HasPrefix(tok, "~"):
		v, err := Parse(tok[1:])
		if err != nil {
			return nil, ErrInvalidRange
		}
		return tildeComparators(v), nil
	case strings.HasPrefix(tok, ">="):
		return oneComparator(opGE, tok[2:])
	case strings.HasPrefix(tok, "<="):
		return oneComparator(opLE, tok[2:])
	case strings.HasPrefix(tok, "!="):
		return oneComparator(opNE, tok[2:])
	case strings.HasPrefix(tok, "=="):
		return oneComparator(opEQ, tok[2:])
	case strings.HasPrefix(tok, ">"):
		return oneComparator(opGT, tok[1:])
	case strings.HasPrefix(tok, "<"):
		return oneComparator(opLT, tok[1:])
	case strings.HasPrefix(tok, "="):
		return oneComparator(opEQ, tok[1:])
	default:
		return oneComparator(opEQ, tok)
	}
}

func oneComparator(op rangeOp, verStr string) ([]rangeComparator, error) {
	v, err := Parse(verStr)
	if err != nil {
		return nil, ErrInvalidRange
	}
	return []rangeComparator{{op: op, v: v}}, nil
}

// caretComparators expands ^v into >=v and an upper bound just past
// the leftmost nonzero component of v (or the patch, if v is 0.0.x).
func caretComparators(v Version) []rangeComparator {
	var upper Version
	switch {
	case v.Major > 0:
		upper = Version{Major: v.Major + 1}
	case v.Minor > 0:
		upper = Version{Minor: v.Minor + 1}
	default:
		upper = Version{Patch: v.Patch + 1}
	}
	return []rangeComparator{
		{op: opGE, v: v},
		{op: opLT, v: upper},
	}
}

// tildeComparators expands ~v into >=v and an upper bound one minor
// version up, leaving the patch component free to vary.
func tildeComparators(v Version) []rangeComparator {
	return []rangeComparator{
		{op: opGE, v: v},
		{op: opLT, v: Version{Major: v.Major, Minor: v.Minor + 1}},
	}
}

// Matches reports whether v satisfies the range: it must satisfy
// every comparator in at least one of the range's comparator sets.
func (r Range) Matches(v Version) bool {
	for _, set := range r.sets {
		if matchesSet(v, set) {
			return true
		}
	}
	return false
}

func matchesSet(v Version, set []rangeComparator) bool {
	for _, c := range set {
		if !matchesComparator(v, c) {
			return false
		}
	}
	return true
}

func matchesComparator(v Version, c rangeComparator) bool {
	cmp := v.Compare(c.v)
	switch c.op {
	case opEQ:
		return cmp == 0
	case opNE:
		return cmp != 0
	case opLT:
		return cmp < 0
	case opLE:
		return cmp <= 0
	case opGT:
		return cmp > 0
	case opGE:
		return cmp >= 0
	default:
		return false
	}
}
