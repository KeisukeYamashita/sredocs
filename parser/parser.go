package parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type Parser interface {
	CompileRegex(fields []string) ([]*regexp.Regexp, error)
	Parse(fields []string, b []byte) (string, error)
	CSVHeader(regexps []*regexp.Regexp) []string
	NamedGroup(field string) string
}

type DefaultParser struct {
}

// CompileRegex compiles regexes based on field names which may include a colon.
func (p *DefaultParser) CompileRegex(fields []string) ([]*regexp.Regexp, error) {
	var r []*regexp.Regexp
	for i, f := range fields {
		/*
			var nextField string
			if i == len(fields)-1 {
				nextField = ""
			} else {
				nextField = fields[i+1]
			}
		*/
		fieldName := p.NamedGroup(fields[i])
		// TODO(stratus): This is the foundation for possibly two
		// regexes - one for easy single line fields and another one for
		// multi-field.
		re, err := regexp.Compile(fmt.Sprintf(`(?mis)%s\s*(?P<%s>.*?)\n`, f, fieldName))
		//re, err := regexp.Compile(fmt.Sprintf(`(?mis)%s\s*(?P<%s>.*?)%s`, f, fieldName, nextField))
		if err != nil {
			return nil, err
		}
		r = append(r, re)
	}
	return r, nil
}

// Parse parses fields out of a slice of bytes into CSV.
func (p *DefaultParser) Parse(fields []string, b []byte) (string, error) {
	regexps, err := p.CompileRegex(fields)
	if err != nil {
		return "", err
	}

	h := p.CSVHeader(regexps)
	records := [][]string{
		h,
	}
	var f []string
	for _, r := range regexps {
		m := r.FindSubmatch(b)
		if len(m) != 2 {
			log.Printf("Could not match regex %s\n", r.String())
			f = append(f, "\"\"")
			continue
		}
		log.Printf("Matched %#v", strings.TrimSpace(string(m[1])))
		f = append(f, strings.TrimSpace(string(m[1])))
	}
	records = append(records, f)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Println(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		log.Println(err)
	}
	return buf.String(), nil
}

// CSVHeader builds a slice out of named groups within a list of regexes.
func (p *DefaultParser) CSVHeader(regexps []*regexp.Regexp) []string {
	var headers []string
	for _, r := range regexps {
		headers = append(headers, r.SubexpNames()[1])
	}
	return headers
}

// NamedGroup converts a field into a name that can be used in a named group.
func (p *DefaultParser) NamedGroup(field string) string {
	r := strings.NewReplacer(" ", "", ":", "", ",", "")
	return r.Replace(strings.ToLower(field))
}
