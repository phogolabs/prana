package sqlexec

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var rgxp = regexp.MustCompile("^\\s*--\\s*name:\\s*(\\S+)")

// Scanner loads a SQL statements for given SQL Script
type Scanner struct{}

// Scan scans a reader for SQL commands that have name tag
func (s *Scanner) Scan(reader io.Reader) map[string]string {
	queries := make(map[string]string)
	scanner := bufio.NewScanner(reader)
	name := ""

	for scanner.Scan() {
		line := scanner.Text()

		if tag := s.tag(line); tag != "" {
			name = tag
		} else if name != "" {
			s.add(name, queries, line)
		}
	}

	return queries
}

func (s *Scanner) tag(line string) string {
	matches := rgxp.FindStringSubmatch(line)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func (s *Scanner) add(name string, queries map[string]string, line string) {
	current := queries[name]
	line = strings.Trim(line, " \t")

	if len(line) == 0 {
		return
	}

	if len(current) > 0 {
		current = current + "\n"
	}

	current = current + line
	queries[name] = current
}
