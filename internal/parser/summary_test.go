package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSimpleList(t *testing.T) {
	summary := `# Summary

- [Chapter 1](ch1.md)
- [Chapter 2](ch2.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	assert.Len(t, s.NumberedChapters, 2)
	assert.Equal(t, "Chapter 1", s.NumberedChapters[0].Title)
	assert.Equal(t, "ch1.md", *s.NumberedChapters[0].Location)
	assert.Equal(t, "Chapter 2", s.NumberedChapters[1].Title)
	assert.Equal(t, "ch2.md", *s.NumberedChapters[1].Location)
}

func TestParseNested(t *testing.T) {
	summary := `# Summary

- [Chapter 1](ch1.md)
  - [Section 1.1](ch1_1.md)
  - [Section 1.2](ch1_2.md)
- [Chapter 2](ch2.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	assert.Len(t, s.NumberedChapters, 2)

	ch1 := s.NumberedChapters[0]
	assert.Equal(t, "Chapter 1", ch1.Title)
	assert.Len(t, ch1.NestedItems, 2)
	assert.Equal(t, "Section 1.1", ch1.NestedItems[0].Title)
	assert.Equal(t, "ch1_1.md", *ch1.NestedItems[0].Location)
}

func TestParseDeepNesting(t *testing.T) {
	summary := `# Summary

- [Ch 1](ch1.md)
  - [S 1.1](s1.md)
    - [S 1.1.1](s1_1.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	ch1 := s.NumberedChapters[0]
	s11 := ch1.NestedItems[0]
	s111 := s11.NestedItems[0]

	assert.Equal(t, "S 1.1.1", s111.Title)
	assert.Equal(t, "s1_1.md", *s111.Location)
}

func TestParseDraftChapter(t *testing.T) {
	summary := `# Summary

- [Intro](intro.md)
- [TODO Chapter]()
- [Conclusion](conclusion.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	assert.Len(t, s.NumberedChapters, 3)

	draft := s.NumberedChapters[1]
	assert.Equal(t, "TODO Chapter", draft.Title)
	assert.Nil(t, draft.Location)
}

func TestParseSeparator(t *testing.T) {
	summary := `# Summary

- [Intro](intro.md)

---

- [Appendix](appendix.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	assert.Len(t, s.PrefixChapters, 1)
	assert.Equal(t, "Intro", s.PrefixChapters[0].Title)

	assert.Len(t, s.SuffixChapters, 1)
	assert.Equal(t, "Appendix", s.SuffixChapters[0].Title)
}

func TestParsePrefixAndSuffix(t *testing.T) {
	summary := `# Summary

- [Foreword](foreword.md)

- [Chapter 1](ch1.md)
- [Chapter 2](ch2.md)

---

- [Appendix](appendix.md)
- [FAQ](faq.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	assert.Len(t, s.PrefixChapters, 1)
	assert.Equal(t, "Foreword", s.PrefixChapters[0].Title)

	assert.Len(t, s.NumberedChapters, 2)

	assert.Len(t, s.SuffixChapters, 2)
	assert.Equal(t, "Appendix", s.SuffixChapters[0].Title)
	assert.Equal(t, "FAQ", s.SuffixChapters[1].Title)
}

func TestValidateSummaryEmpty(t *testing.T) {
	s := &Summary{
		PrefixChapters:   make([]*SummaryItem, 0),
		NumberedChapters: make([]*SummaryItem, 0),
		SuffixChapters:   make([]*SummaryItem, 0),
	}

	err := ValidateSummaryStructure(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no chapters")
}

func TestValidateSummaryValid(t *testing.T) {
	summary := `- [Ch 1](ch1.md)`
	s, _ := ParseSummary(summary)

	err := ValidateSummaryStructure(s)
	assert.NoError(t, err)
}

func TestFlattenSummary(t *testing.T) {
	summary := `# Summary

- [Intro](intro.md)

- [Chapter 1](ch1.md)
- [Chapter 2](ch2.md)

---

- [Appendix](appendix.md)
`

	s, _ := ParseSummary(summary)
	flat := s.FlattenSummary()

	assert.Len(t, flat, 5)
	assert.Equal(t, "Intro", flat[0].Title)
	assert.Equal(t, "Chapter 1", flat[1].Title)
	assert.Equal(t, "Appendix", flat[4].Title)
}

func TestParsePathsWithDirs(t *testing.T) {
	summary := `# Summary

- [Chapter 1](dir1/ch1.md)
  - [Section](dir1/sec.md)
- [Chapter 2](dir2/subdir/ch2.md)
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	ch1 := s.NumberedChapters[0]
	assert.Equal(t, "dir1/ch1.md", *ch1.Location)
	assert.Equal(t, "dir1/sec.md", *ch1.NestedItems[0].Location)

	ch2 := s.NumberedChapters[1]
	assert.Equal(t, "dir2/subdir/ch2.md", *ch2.Location)
}

func TestParseEmptySummary(t *testing.T) {
	summary := `# Summary

# Just headers
No list items here
`

	s, err := ParseSummary(summary)
	require.NoError(t, err)

	assert.Len(t, s.NumberedChapters, 0)
}
