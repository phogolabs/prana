package sqlexec

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
)

var separatorRgxp = regexp.MustCompile(`^[\s]*[-]*[\s]*(?i)go[;]*\s*`)

// Splitter splits a statement by GO separator
type Splitter struct{}

// Split splits a statement by GO separator
func (s *Splitter) Split(reader io.Reader) []string {
	buffer := &bytes.Buffer{}
	queries := []string{}
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if s.match(line) {
			s.add(buffer, &queries)
			continue
		}

		fmt.Fprintln(buffer, line)
	}

	s.add(buffer, &queries)
	return queries
}

func (s *Splitter) match(line string) bool {
	return separatorRgxp.MatchString(line)
}

func (s *Splitter) add(buffer *bytes.Buffer, queries *[]string) {
	if buffer.Len() > 0 {
		*queries = append(*queries, buffer.String())
		buffer.Reset()
	}
}
