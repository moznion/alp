package stats

import (
	"fmt"
	"io"
	"strings"

	"github.com/tkuchiki/alp/helpers"

	"github.com/olekukonko/tablewriter"
)

var headers = map[string]string{
	"count":              "Count",
	"1xx":                "1xx",
	"2xx":                "2xx",
	"3xx":                "3xx",
	"4xx":                "4xx",
	"5xx":                "5xx",
	"method":             "Method",
	"uri":                "Uri",
	"min":                "Min",
	"max":                "Max",
	"sum":                "Sum",
	"avg":                "Avg",
	"p1":                 "P1",
	"p50":                "P50",
	"p99":                "P99",
	"stddev":             "Stddev",
	"min_body":           "Min(Body)",
	"max_body":           "Max(Body)",
	"sum_body":           "Sum(Body)",
	"avg_body":           "Avg(Body)",
	"breakdown_statuses": "Statuses Breakdown",
}

var keywords = []string{
	"count",
	"1xx",
	"2xx",
	"3xx",
	"4xx",
	"5xx",
	"method",
	"uri",
	"min",
	"max",
	"sum",
	"avg",
	"p1",
	"p50",
	"p99",
	"stddev",
	"min_body",
	"max_body",
	"sum_body",
	"avg_body",
	"breakdown_statuses",
}

var defaultHeaders = []string{
	"Count", "1xx", "2xx", "3xx", "4xx", "5xx", "Method", "Uri",
	"Min", "Max", "Sum", "Avg",
	"P1", "P50", "P99", "Stddev",
	"Min(Body)", "Max(Body)", "Sum(Body)", "Avg(Body)",
}

type Printer struct {
	keywords          []string
	format            string
	noHeaders         bool
	showFooters       bool
	breakdownStatuses bool
	headers           []string
	writer            io.Writer
	all               bool
}

func NewPrinter(w io.Writer, val, format string, noHeaders, showFooters, breakdownStatuses bool) *Printer {
	p := &Printer{
		format:            format,
		writer:            w,
		showFooters:       showFooters,
		noHeaders:         noHeaders,
		breakdownStatuses: breakdownStatuses,
	}

	if val == "all" {
		p.keywords = keywords
		p.headers = defaultHeaders
		p.all = true
	} else {
		p.keywords = helpers.SplitCSV(val)
		for _, key := range p.keywords {
			p.headers = append(p.headers, headers[key])
			if key == "all" {
				p.keywords = keywords
				p.headers = defaultHeaders
				p.all = true
				break
			}
		}
	}

	if p.breakdownStatuses {
		p.headers = append(p.headers, headers["breakdown_statuses"])
	}

	return p
}

func (p *Printer) Validate() error {
	if p.all {
		return nil
	}

	invalids := make([]string, 0)
	for _, key := range p.keywords {
		if _, ok := headers[key]; !ok {
			invalids = append(invalids, key)
		}
	}

	if len(invalids) > 0 {
		return fmt.Errorf("invalid keywords: %s", strings.Join(invalids, ","))
	}

	return nil
}

func (p *Printer) generateAllLine(s *HTTPStat, quoteUri bool) []string {
	uri := s.Uri
	if quoteUri && strings.Contains(s.Uri, ",") {
		uri = fmt.Sprintf(`"%s"`, s.Uri)
	}

	cols := []string{
		s.StrCount(),
		s.StrStatus1xx(),
		s.StrStatus2xx(),
		s.StrStatus3xx(),
		s.StrStatus4xx(),
		s.StrStatus5xx(),
		s.Method,
		uri,
		round(s.MinResponseTime()),
		round(s.MaxResponseTime()),
		round(s.SumResponseTime()),
		round(s.AvgResponseTime()),
		round(s.P1ResponseTime()),
		round(s.P50ResponseTime()),
		round(s.P99ResponseTime()),
		round(s.StddevResponseTime()),
		round(s.MinResponseBodyBytes()),
		round(s.MaxResponseBodyBytes()),
		round(s.SumResponseBodyBytes()),
		round(s.AvgResponseBodyBytes()),
	}

	if p.breakdownStatuses {
		cols = append(cols, s.StrCountsForEachStatus())
	}

	return cols
}

func (p *Printer) GenerateLine(s *HTTPStat, quoteUri bool) []string {
	if p.all {
		return p.generateAllLine(s, quoteUri)
	}

	keyLen := len(p.keywords)
	line := make([]string, 0, keyLen)

	for i := 0; i < keyLen; i++ {
		switch p.keywords[i] {
		case "count":
			line = append(line, s.StrCount())
		case "method":
			line = append(line, s.Method)
		case "uri":
			uri := s.Uri
			if quoteUri && strings.Contains(s.Uri, ",") {

				uri = fmt.Sprintf(`"%s"`, s.Uri)
			}
			line = append(line, uri)
		case "1xx":
			line = append(line, s.StrStatus1xx())
		case "2xx":
			line = append(line, s.StrStatus2xx())
		case "3xx":
			line = append(line, s.StrStatus3xx())
		case "4xx":
			line = append(line, s.StrStatus4xx())
		case "5xx":
			line = append(line, s.StrStatus5xx())
		case "min":
			line = append(line, round(s.MinResponseTime()))
		case "max":
			line = append(line, round(s.MaxResponseTime()))
		case "sum":
			line = append(line, round(s.SumResponseTime()))
		case "avg":
			line = append(line, round(s.AvgResponseTime()))
		case "p1":
			line = append(line, round(s.P1ResponseTime()))
		case "p50":
			line = append(line, round(s.P50ResponseTime()))
		case "p99":
			line = append(line, round(s.P99ResponseTime()))
		case "stddev":
			line = append(line, round(s.StddevResponseTime()))
		case "min_body":
			line = append(line, round(s.MinResponseBodyBytes()))
		case "max_body":
			line = append(line, round(s.MaxResponseBodyBytes()))
		case "sum_body":
			line = append(line, round(s.SumResponseBodyBytes()))
		case "avg_body":
			line = append(line, round(s.AvgResponseBodyBytes()))
		case "breakdown_statuses":
			line = append(line, s.StrCountsForEachStatus())
		}
	}

	return line
}

func (p *Printer) GenerateFooter(counts map[string]int) []string {
	keyLen := len(p.keywords)
	line := make([]string, 0, keyLen)

	for i := 0; i < keyLen; i++ {
		switch p.keywords[i] {
		case "count":
			line = append(line, fmt.Sprint(counts["count"]))
		case "1xx":
			line = append(line, fmt.Sprint(counts["1xx"]))
		case "2xx":
			line = append(line, fmt.Sprint(counts["2xx"]))
		case "3xx":
			line = append(line, fmt.Sprint(counts["3xx"]))
		case "4xx":
			line = append(line, fmt.Sprint(counts["4xx"]))
		case "5xx":
			line = append(line, fmt.Sprint(counts["5xx"]))
		default:
			line = append(line, "")
		}
	}

	return line
}

func (p *Printer) SetFormat(format string) {
	p.format = format
}

func (p *Printer) SetHeaders(headers []string) {
	p.headers = headers
}

func (p *Printer) SetWriter(w io.Writer) {
	p.writer = w
}

func (p *Printer) Print(hs *HTTPStats) {
	switch p.format {
	case "table":
		p.printTable(hs)
	case "md", "markdown":
		p.printMarkdown(hs)
	case "tsv":
		p.printTSV(hs)
	case "csv":
		p.printCSV(hs)
	}
}

func round(num float64) string {
	return fmt.Sprintf("%.3f", num)
}

func (p *Printer) printTable(hs *HTTPStats) {
	table := tablewriter.NewWriter(p.writer)
	table.SetHeader(p.headers)
	for _, s := range hs.stats {
		data := p.GenerateLine(s, false)
		table.Append(data)
	}

	if p.showFooters {
		footer := p.GenerateFooter(hs.CountAll())
		table.SetFooter(footer)
		table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)
	}

	table.Render()
}

func (p *Printer) printMarkdown(hs *HTTPStats) {
	table := tablewriter.NewWriter(p.writer)
	table.SetHeader(p.headers)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, s := range hs.stats {
		data := p.GenerateLine(s, false)
		table.Append(data)
	}

	if p.showFooters {
		footer := p.GenerateFooter(hs.CountAll())
		table.Append(footer)
	}

	table.Render()
}

func (p *Printer) printTSV(hs *HTTPStats) {
	if !p.noHeaders {
		fmt.Println(strings.Join(p.headers, "\t"))
	}
	for _, s := range hs.stats {
		data := p.GenerateLine(s, false)
		fmt.Println(strings.Join(data, "\t"))
	}
}

func (p *Printer) printCSV(hs *HTTPStats) {
	if !p.noHeaders {
		fmt.Println(strings.Join(p.headers, ","))
	}
	for _, s := range hs.stats {
		data := p.GenerateLine(s, true)
		fmt.Println(strings.Join(data, ","))
	}
}
