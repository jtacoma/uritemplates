package uritemplates

import (
	"errors"
	"net/url"
	"strings"
)

type UriTemplate struct {
	raw   string
	parts []templatePart
}

const (
	_          = iota
	LEVEL1     = iota
	PLUS       = iota
	CROSSHATCH = iota
	SLASH      = iota
	DOT        = iota
	SEMICOLON  = iota
	QUERY      = iota
	AMPERSAND  = iota
)

type templatePart struct {
	raw   string
	kind  int
	terms []string
}

func Parse(rawtemplate string) (template *UriTemplate, err error) {
	template = new(UriTemplate)
	template.raw = rawtemplate
	template.parts = make([]templatePart, 32) // TODO: smart allocation
	split := strings.Split(rawtemplate, "{")
	for i, s := range split {
		if i == 0 {
			template.parts[i].raw = s
		} else {
			subsplit := strings.Split(s, "}")
			if len(subsplit) != 2 {
				return nil, errors.New("malformed template")
			}
			expression := subsplit[0]
			terms := make([]string, 32) // TODO: smart allocation
			terms[0] = expression
			template.parts[i*2-1].kind = LEVEL1
			template.parts[i*2-1].terms = terms
			template.parts[i*2].raw = subsplit[1]
		}
	}
	return template, nil
}

func (self *UriTemplate) ExpandString(values map[string]string) (string, error) {
	raw := ""
	for _, p := range self.parts {
		if len(p.raw) > 0 {
			raw = raw + p.raw
		} else if p.kind == LEVEL1 {
			for _, term := range p.terms {
				if value, ok := values[term]; ok {
					raw = raw + value
				}
			}
		}
	}
	return raw, nil
}

func (self *UriTemplate) ExpandURL(values map[string]string) (expanded *url.URL, err error) {
	var s string
	if s, err = self.ExpandString(values); err != nil {
		return expanded, err
	}
	return url.Parse(s)
}
