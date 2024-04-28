package yaml

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aryann/difflib"
	"github.com/fatih/color"
)

func MergeDiffOptions(opts []DiffOptions) DiffOptions {
	var merged DiffOptions
	for _, v := range opts {
		if v.ContextLineN > merged.ContextLineN {
			merged.ContextLineN = v.ContextLineN
		}
	}
	return merged
}

type DiffOptions struct {
	ContextLineN int
}

func (o *DiffOptions) Diff(x, y string) string {
	var (
		sb   strings.Builder
		curr sequence = sequence{}
	)
	diffs := difflib.Diff(strings.Split(x, "\n"), strings.Split(y, "\n"))
	curr.reset(diffs, 0)

	for i, v := range diffs {
		if o.ContextLineN < 1 {
			// add all records
			sb.WriteString(printDiff(v))
			continue
		}

		curr.incrementLineN(v)

		// check if current cursor is divider
		if divExp.MatchString(v.Payload) {
			// if diff sequence is in progress, stop it
			if curr.isDiffSequence {
				curr.stopDiff()

				// add divider
				sb.WriteString("\n")
			}

			// reset cursor for next yaml sequence
			// current cursor is divider so reset with next index
			curr.reset(diffs, i+1)

			continue
		}

		log().Debug("processing diff line", "index", i, "lineN", curr.fileLineN, "kind", curr.kind,
			"name", curr.name, "preIsDiff", curr.isDiffSequence, "delta", v.Delta, "payload", v.Payload)

		if !curr.isDiffSequence {
			if v.Delta != difflib.Common {
				curr.recordDiff()

				// add header
				sb.WriteString(curr.header())

				// add previous lines
				for j := numInRange(curr.startIndex, len(diffs), i-o.ContextLineN); j < i; j++ {
					if divExp.MatchString(diffs[j].Payload) {
						continue
					}
					sb.WriteString(fmt.Sprintf("%s\n", diffs[j]))
				}

				sb.WriteString(printDiff(v))
			}

		} else {
			// is diff sequence
			sb.WriteString(printDiff(v))

			newDiffFound := false
			for j := i + 1; j < numInRange(0, len(diffs), i+o.ContextLineN+1); j++ {
				if diffs[j].Delta != difflib.Common {
					newDiffFound = true
					break
				}
			}
			if !newDiffFound {
				curr.stopDiff()

				// add subsequent lines
				for j := i + 1; j < numInRange(0, len(diffs), i+o.ContextLineN+1); j++ {
					if divExp.MatchString(diffs[j].Payload) {
						break
					}
					sb.WriteString(fmt.Sprintf("%s\n", diffs[j]))
				}
				// add divider
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

var (
	divExp  = regexp.MustCompile(`^---$`)
	kindExp = regexp.MustCompile(`^kind: (.+)$`)
	metaExp = regexp.MustCompile(`^metadata:$`)
	nameExp = regexp.MustCompile(`^  name: (.+)$`)
)

type sequence struct {
	kind           string
	name           string
	startIndex     int
	isDiffSequence bool

	fileLineN int
}

func (s *sequence) incrementLineN(r difflib.DiffRecord) {
	// record actual snapshot file line number
	// right only means new line added so it should not be considered
	if r.Delta != difflib.RightOnly {
		s.fileLineN++
	}
}

func (s *sequence) reset(diffs []difflib.DiffRecord, i int) {
	s.kind, s.name = findNextKind(diffs[i:]), findNextName(diffs[i:])
	s.startIndex = i
	s.isDiffSequence = false
}

func (s *sequence) recordDiff() {
	s.isDiffSequence = true
}

func (s *sequence) stopDiff() {
	s.isDiffSequence = false
}

func (s *sequence) header() string {
	return printHeader(s.kind, s.name, s.fileLineN)
}

func numInRange(min, max, v int) int {
	if v >= min && v <= max {
		return v
	} else if v < min {
		return min
	} else {
		return max
	}
}

func printDiff(d difflib.DiffRecord) string {
	switch d.Delta {
	case difflib.LeftOnly:
		return color.New(color.FgRed).Sprintf("%s\n", d)
	case difflib.RightOnly:
		return color.New(color.FgGreen).Sprintf("%s\n", d)
	default:
		return fmt.Sprintf("%s\n", d)
	}
}

func printHeader(kind, name string, lineN int) string {
	return color.New(color.FgCyan, color.Bold, color.Italic).Sprintf("@@ KIND=%s NAME=%s LINE=%d\n", kind, name, lineN)
}

func findNextKind(diffs []difflib.DiffRecord) string {
	for i := 0; i < len(diffs); i++ {
		kindMatch := kindExp.FindStringSubmatch(diffs[i].Payload)
		if len(kindMatch) > 0 {
			return kindMatch[1]
		}
	}
	return ""
}

func findNextName(diffs []difflib.DiffRecord) string {
	for i := 0; i < len(diffs); i++ {
		if metaExp.MatchString(diffs[i].Payload) {
			for j := i + 1; j < len(diffs)-i; j++ {
				nameMatch := nameExp.FindStringSubmatch(diffs[j].Payload)
				if len(nameMatch) > 0 {
					return nameMatch[1]
				}
			}
			return ""
		}
	}
	return ""
}
