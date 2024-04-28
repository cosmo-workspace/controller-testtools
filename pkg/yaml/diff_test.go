package yaml

import (
	"os"
	"testing"

	. "github.com/jlandowner/helm-chartsnap/pkg/snap/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/aryann/difflib"
)

var _ = Describe("Diff", func() {
	Context("DiffContextLineN is 3", func() {
		It("should return the extracted diff with previous/next 3 lines", func() {
			expectedSnap := mustReadFile("testdata/expected.snap")
			actualSnap := mustReadFile("testdata/actual.snap")

			d := DiffOptions{
				ContextLineN: 3,
			}
			diff := d.Diff(expectedSnap, actualSnap)
			Ω(diff).To(MatchSnapShot())
		})
	})

	Context("DiffContextLineN is 0", func() {
		It("should return all diff", func() {
			expectedSnap := mustReadFile("testdata/expected.snap")
			actualSnap := mustReadFile("testdata/actual.snap")

			d := DiffOptions{
				ContextLineN: 0,
			}
			diff := d.Diff(expectedSnap, actualSnap)
			Ω(diff).To(MatchSnapShot())
		})
	})
})

func TestMergeDiffOptions(t *testing.T) {
	opts := []DiffOptions{
		{ContextLineN: 5},
		{ContextLineN: 3},
		{ContextLineN: 7},
	}

	merged := MergeDiffOptions(opts)

	if merged.ContextLineN != 7 {
		t.Errorf("Expected DiffContextLineN to be 7, got %d", merged.ContextLineN)
	}
}

func Test_printDiff(t *testing.T) {
	d := difflib.DiffRecord{
		Delta:   difflib.LeftOnly,
		Payload: "abc",
	}

	want := "- abc\n"

	if got := printDiff(d); got != want {
		t.Errorf("printDiff() = %v, want %v", got, want)
	}
}

func Test_printHeader(t *testing.T) {
	kind := "TestKind"
	name := "TestName"
	lineN := 5

	want := "@@ KIND=TestKind NAME=TestName LINE=5\n"

	if got := printHeader(kind, name, lineN); got != want {
		t.Errorf("printHeader() = %v, want %v", got, want)
	}
}

func Test_findNextKind(t *testing.T) {
	diffs := []difflib.DiffRecord{
		{Delta: difflib.Common, Payload: "abc"},
		{Delta: difflib.Common, Payload: "kind: Pod"},
		{Delta: difflib.Common, Payload: "def"},
	}

	want := "Pod"

	if got := findNextKind(diffs); got != want {
		t.Errorf("findNextKind() = %v, want %v", got, want)
	}
}

func Test_findNextName(t *testing.T) {
	diffs := []difflib.DiffRecord{
		{Delta: difflib.Common, Payload: "abc"},
		{Delta: difflib.Common, Payload: "metadata:"},
		{Delta: difflib.Common, Payload: "  name: TestName"},
		{Delta: difflib.Common, Payload: "def"},
	}

	want := "TestName"

	if got := findNextName(diffs); got != want {
		t.Errorf("findNextName() = %v, want %v", got, want)
	}
}

func mustReadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}
