package web

import (
	"html/template"
	"strings"
)

// mdLite renders the small markdown subset used by content/*.md: headings,
// bullets, pipe tables, **bold**, `code`, and paragraphs. It exists so the
// station/resource pages read like documentation instead of raw source.
//
// It escapes everything first, so content files can never inject HTML.
func mdLite(src string) template.HTML {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	lines := strings.Split(src, "\n")

	var out strings.Builder
	var listOpen, tableOpen bool
	var tableRows [][]string

	closeList := func() {
		if listOpen {
			out.WriteString("</ul>")
			listOpen = false
		}
	}
	closeTable := func() {
		if !tableOpen {
			return
		}
		out.WriteString("<table>")
		for i, row := range tableRows {
			out.WriteString("<tr>")
			for _, cell := range row {
				if i == 0 {
					out.WriteString("<th>" + inline(cell) + "</th>")
				} else {
					out.WriteString("<td>" + inline(cell) + "</td>")
				}
			}
			out.WriteString("</tr>")
		}
		out.WriteString("</table>")
		tableRows = nil
		tableOpen = false
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, " ")
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "":
			closeList()
			closeTable()
			continue

		case strings.HasPrefix(trimmed, "### "):
			closeList()
			closeTable()
			out.WriteString("<h4>" + inline(trimmed[4:]) + "</h4>")

		case strings.HasPrefix(trimmed, "## "):
			closeList()
			closeTable()
			out.WriteString("<h3>" + inline(trimmed[3:]) + "</h3>")

		case strings.HasPrefix(trimmed, "# "):
			closeList()
			closeTable()
			out.WriteString("<h2>" + inline(trimmed[2:]) + "</h2>")

		case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "):
			closeTable()
			if !listOpen {
				out.WriteString("<ul>")
				listOpen = true
			}
			out.WriteString("<li>" + inline(trimmed[2:]) + "</li>")

		case strings.HasPrefix(trimmed, "|"):
			closeList()
			if !tableOpen {
				tableOpen = true
			}
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}
			// A separator row (---) carries no content.
			if isSeparator(cells) {
				continue
			}
			tableRows = append(tableRows, cells)

		default:
			closeList()
			closeTable()
			out.WriteString("<p>" + inline(trimmed) + "</p>")
		}
	}
	closeList()
	closeTable()
	return template.HTML(out.String())
}

func isSeparator(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

// inline handles the two inline constructs we allow. Input is already escaped
// by escapeHTML, so the replacements only introduce our own tags.
func inline(s string) string {
	s = template.HTMLEscapeString(s)
	s = strings.ReplaceAll(s, "`", "")
	// **bold**
	for {
		i := strings.Index(s, "**")
		if i < 0 {
			break
		}
		rest := s[i+2:]
		j := strings.Index(rest, "**")
		if j < 0 {
			s = s[:i] + rest
			break
		}
		s = s[:i] + "<strong>" + rest[:j] + "</strong>" + rest[j+2:]
	}
	return s
}