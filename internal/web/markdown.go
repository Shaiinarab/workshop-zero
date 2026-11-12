package web

import (
	"html/template"
	"regexp"
	"strings"
)

var (
	linkPattern = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)\s]+)\)`)
	codePattern = regexp.MustCompile("`([^`]+)`")
)

// mdLite renders the small, safe Markdown subset used by content/*.md.
// It supports headings, ordered/unordered lists, blockquotes, pipe tables,
// links, bold, inline code and paragraphs. Raw HTML is always escaped.
func mdLite(src string) template.HTML {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	lines := strings.Split(src, "\n")
	var out strings.Builder
	var listTag string
	var tableOpen bool
	var tableRows [][]string

	closeList := func() {
		if listTag != "" {
			out.WriteString("</" + listTag + ">")
			listTag = ""
		}
	}
	closeTable := func() {
		if !tableOpen { return }
		out.WriteString("<table>")
		for i, row := range tableRows {
			out.WriteString("<tr>")
			for _, cell := range row {
				if i == 0 { out.WriteString("<th>" + inline(cell) + "</th>") } else { out.WriteString("<td>" + inline(cell) + "</td>") }
			}
			out.WriteString("</tr>")
		}
		out.WriteString("</table>")
		tableRows = nil
		tableOpen = false
	}

	for _, raw := range lines {
		trimmed := strings.TrimSpace(strings.TrimRight(raw, " "))
		if trimmed == "" { closeList(); closeTable(); continue }
		switch {
		case strings.HasPrefix(trimmed, "### "):
			closeList(); closeTable(); out.WriteString("<h4>" + inline(trimmed[4:]) + "</h4>")
		case strings.HasPrefix(trimmed, "## "):
			closeList(); closeTable(); out.WriteString("<h3>" + inline(trimmed[3:]) + "</h3>")
		case strings.HasPrefix(trimmed, "# "):
			closeList(); closeTable(); out.WriteString("<h2>" + inline(trimmed[2:]) + "</h2>")
		case strings.HasPrefix(trimmed, "> "):
			closeList(); closeTable(); out.WriteString("<blockquote><p>" + inline(trimmed[2:]) + "</p></blockquote>")
		case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "):
			closeTable()
			if listTag != "ul" { closeList(); out.WriteString("<ul>"); listTag = "ul" }
			out.WriteString("<li>" + inline(trimmed[2:]) + "</li>")
		case isOrderedItem(trimmed):
			closeTable()
			if listTag != "ol" { closeList(); out.WriteString("<ol>"); listTag = "ol" }
			idx := strings.Index(trimmed, ". ")
			out.WriteString("<li>" + inline(trimmed[idx+2:]) + "</li>")
		case strings.HasPrefix(trimmed, "|"):
			closeList()
			if !tableOpen { tableOpen = true }
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			for i := range cells { cells[i] = strings.TrimSpace(cells[i]) }
			if !isSeparator(cells) { tableRows = append(tableRows, cells) }
		case strings.Trim(trimmed, "-") == "" && len(trimmed) >= 3:
			closeList(); closeTable(); out.WriteString("<hr>")
		default:
			closeList(); closeTable(); out.WriteString("<p>" + inline(trimmed) + "</p>")
		}
	}
	closeList(); closeTable()
	return template.HTML(out.String())
}

func isOrderedItem(s string) bool {
	i := strings.Index(s, ". ")
	if i < 1 || i > 3 { return false }
	for _, r := range s[:i] { if r < '0' || r > '9' { return false } }
	return true
}

func isSeparator(cells []string) bool {
	if len(cells) == 0 { return false }
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c == "" { continue }
		if strings.Trim(c, "-: ") != "" { return false }
	}
	return true
}

func inline(s string) string {
	s = template.HTMLEscapeString(s)
	s = codePattern.ReplaceAllString(s, "<code>$1</code>")
	for {
		i := strings.Index(s, "**")
		if i < 0 { break }
		rest := s[i+2:]
		j := strings.Index(rest, "**")
		if j < 0 { s = s[:i] + rest; break }
		s = s[:i] + "<strong>" + rest[:j] + "</strong>" + rest[j+2:]
	}
	s = linkPattern.ReplaceAllString(s, `<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>`)
	return s
}
