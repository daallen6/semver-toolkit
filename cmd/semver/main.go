// Command semver is a thin wrapper around the semver package for
// validating, sorting, and comparing version strings from the shell.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/daallen6/semver-toolkit"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var code int
	switch os.Args[1] {
	case "validate":
		code = runValidate(os.Args[2:])
	case "sort":
		code = runSort(os.Args[2:])
	case "compare":
		code = runCompare(os.Args[2:])
	case "-h", "--help", "help":
		usage()
		code = 0
	default:
		fmt.Fprintf(os.Stderr, "semver: unknown command %q\n\n", os.Args[1])
		usage()
		code = 2
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: semver <command> [arguments]

commands:
  validate [files...]   check each line is a valid semantic version
  sort [files...]       print versions in ascending order
  compare <a> <b>       print <, =, or > comparing two versions

validate and sort read from the given files, or from stdin if no
files are given. Blank lines and lines starting with # are ignored.`)
}

func runValidate(paths []string) int {
	lines, err := readLines(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "semver:", err)
		return 1
	}
	bad := 0
	for _, line := range lines {
		if _, err := semver.Parse(line); err != nil {
			fmt.Printf("INVALID %s\n", line)
			bad++
			continue
		}
		fmt.Printf("OK      %s\n", line)
	}
	if bad > 0 {
		return 1
	}
	return 0
}

func runSort(paths []string) int {
	lines, err := readLines(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "semver:", err)
		return 1
	}
	versions := make([]semver.Version, 0, len(lines))
	for _, line := range lines {
		v, err := semver.Parse(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "semver: skipping invalid version %q: %v\n", line, err)
			continue
		}
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return semver.Less(versions[i], versions[j]) })
	for _, v := range versions {
		fmt.Println(v.String())
	}
	return 0
}

func runCompare(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: semver compare <version> <version>")
		return 2
	}
	a, err := semver.Parse(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "semver: %s: %v\n", args[0], err)
		return 1
	}
	b, err := semver.Parse(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "semver: %s: %v\n", args[1], err)
		return 1
	}
	switch c := a.Compare(b); {
	case c < 0:
		fmt.Println("<")
	case c > 0:
		fmt.Println(">")
	default:
		fmt.Println("=")
	}
	return 0
}

// readLines gathers non-empty, non-comment lines from the given
// paths, or from stdin if paths is empty. A path of "-" also means
// stdin, so it can be mixed with real files.
func readLines(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return scanLines(os.Stdin)
	}
	var lines []string
	for _, p := range paths {
		ls, err := readLinesFromPath(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		lines = append(lines, ls...)
	}
	return lines, nil
}

func readLinesFromPath(p string) ([]string, error) {
	if p == "-" {
		return scanLines(os.Stdin)
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return scanLines(f)
}

func scanLines(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}
