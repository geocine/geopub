package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// SummaryItem represents an item in SUMMARY.md
type SummaryItem struct {
	Type        string // "link", "separator", "part-title"
	Title       string
	Location    *string // Relative path to markdown file
	NestedItems []*SummaryItem
	Number      *SectionNumber
}

// SectionNumber represents section numbering
type SectionNumber struct {
	Parts []int
}

// Summary represents parsed SUMMARY.md
type Summary struct {
    Title            string
    PrefixChapters   []*SummaryItem
    NumberedChapters []*SummaryItem
    SuffixChapters   []*SummaryItem
    HasMiddleSeparator bool
}

// ParseSummary parses SUMMARY.md content and returns a Summary
func ParseSummary(content string) (*Summary, error) {
	summary := &Summary{
		PrefixChapters:   make([]*SummaryItem, 0),
		NumberedChapters: make([]*SummaryItem, 0),
		SuffixChapters:   make([]*SummaryItem, 0),
	}

    lines := strings.Split(content, "\n")
    state := "unknown" // unknown, prefix, numbered, suffix
    var parentStack []*SummaryItem
    // Track first list block to optionally classify as prefix
    var firstTopItem *SummaryItem
    firstBlockTopCount := 0
    totalTopLevel := 0
    seenAnyList := false
    inFirstBlock := false
    seenBlankAfterFirstBlock := false

	// Regex for matching link pattern: [Title](path.md)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]*)\)`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

        // Skip empty lines; also mark end of first block if applicable
        if trimmed == "" {
            if seenAnyList && inFirstBlock && !seenBlankAfterFirstBlock {
                seenBlankAfterFirstBlock = true
            }
            continue
        }

		// Skip markdown headers
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for separator
        if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "***") {
            if state == "unknown" {
                state = "numbered" // First separator: transition unknown -> numbered
            } else if state == "prefix" || state == "numbered" {
                state = "suffix" // Separator after prefix or numbered: transition to suffix
                summary.HasMiddleSeparator = true
            }
            parentStack = nil // Reset parent stack on separator
            continue
        }

		// Calculate indentation level
		indent := len(line) - len(strings.TrimLeft(line, " "))
		level := indent / 2 // Assuming 2 spaces per level

		// Check for list items
        if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
            // Default to numbered for the first block
            if state == "unknown" {
                state = "numbered"
                seenAnyList = true
                inFirstBlock = true
            }

			// Remove the list marker
			content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "-"), "*")
			content = strings.TrimSpace(content)

            // Try to find a link
            matches := linkRegex.FindStringSubmatch(content)
            if len(matches) == 3 {
                title := matches[1]
                location := matches[2]

				item := &SummaryItem{
					Type:        "link",
					Title:       title,
					NestedItems: make([]*SummaryItem, 0),
				}

                if location != "" {
                    item.Location = &location
                }

                // Track number of top-level items in the first block
                if level == 0 {
                    totalTopLevel++
                }
                if inFirstBlock && !seenBlankAfterFirstBlock && level == 0 {
                    firstBlockTopCount++
                    if firstTopItem == nil {
                        firstTopItem = item
                    }
                }

				// Adjust parent stack based on level
				for len(parentStack) > level {
					parentStack = parentStack[:len(parentStack)-1]
				}

				// Determine target list based on state
				var targetList *[]*SummaryItem
				switch state {
				case "prefix":
					targetList = &summary.PrefixChapters
				case "numbered":
					targetList = &summary.NumberedChapters
				case "suffix":
					targetList = &summary.SuffixChapters
				default:
					targetList = &summary.NumberedChapters // Fallback to numbered
				}

                // Add to appropriate list
                if level == 0 {
                    parentStack = []*SummaryItem{item}
                    *targetList = append(*targetList, item)
                } else if len(parentStack) > 0 {
                    parent := parentStack[len(parentStack)-1]
                    parent.NestedItems = append(parent.NestedItems, item)
                    parentStack = append(parentStack, item)
                }
            }
        }
    }

    // Heuristic: if the first block had exactly one top-level item, treat it as prefix
    if firstBlockTopCount == 1 && totalTopLevel > 1 && len(summary.NumberedChapters) > 0 && summary.NumberedChapters[0] == firstTopItem {
        summary.NumberedChapters = summary.NumberedChapters[1:]
        summary.PrefixChapters = append(summary.PrefixChapters, firstTopItem)
    }

	return summary, nil
}

// ValidateSummaryStructure validates the summary structure
func ValidateSummaryStructure(summary *Summary) error {
	// Ensure at least some chapters
	if len(summary.NumberedChapters) == 0 &&
		len(summary.PrefixChapters) == 0 &&
		len(summary.SuffixChapters) == 0 {
		return fmt.Errorf("SUMMARY.md contains no chapters")
	}

	return nil
}

// FlattenSummary returns all items in order: prefix, numbered, suffix
func (s *Summary) FlattenSummary() []*SummaryItem {
    items := make([]*SummaryItem, 0)
    items = append(items, s.PrefixChapters...)
    items = append(items, s.NumberedChapters...)
    if s.HasMiddleSeparator {
        items = append(items, &SummaryItem{Type: "separator", NestedItems: make([]*SummaryItem, 0)})
    }
    items = append(items, s.SuffixChapters...)
    return items
}

// AssignSectionNumbers assigns section numbers to chapters
func (s *Summary) AssignSectionNumbers() {
	// Number all top-level link items in FlattenSummary order
	topItems := s.FlattenSummary()
	topIndex := 0
	for _, item := range topItems {
		if item.Type != "link" {
			continue
		}
		topIndex++
		num := []int{topIndex}
		assignNumbersToItem(item, num)
	}
}

// assignNumbersToItem sets the number on an item and recursively numbers its link children
func assignNumbersToItem(item *SummaryItem, number []int) {
	item.Number = &SectionNumber{Parts: make([]int, len(number))}
	copy(item.Number.Parts, number)

	childIndex := 0
	for _, child := range item.NestedItems {
		if child.Type != "link" {
			continue
		}
		childIndex++
		childNum := append(append([]int{}, number...), childIndex)
		assignNumbersToItem(child, childNum)
	}
}

// Link represents a link entry in SUMMARY.md
type Link struct {
	Name        string
	Location    *string // Can be nil for draft chapters
	Number      *SectionNumber
	NestedItems []*SummaryItem
}
