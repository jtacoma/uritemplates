package uritemplates

import (
	"errors"
	"fmt"
	//"html"
	//"html/template"
	//"net/url"
	"regexp"
	"strings"
)

var unreserved = regexp.MustCompile("[^A-Za-z0-9\\-._~]")
var reserved = regexp.MustCompile("[^A-Za-z0-9\\-._~:/?#[\\]@!$&'()*+,;=]")

func pctEncode(original string) (result string) {
	for _, b := range []byte(original) {
		if b < 16 {
			result += fmt.Sprintf("%%0%X", b)
		} else {
			result += fmt.Sprintf("%%%X", b)
		}
	}
	return result
}

func escape(s string, allowReserved bool) string {
	var result string
	if allowReserved {
		result = reserved.ReplaceAllStringFunc(s, pctEncode)
	} else {
		result = unreserved.ReplaceAllStringFunc(s, pctEncode)
	}
	return result
}

type UriTemplate struct {
	raw   string
	parts []templatePart
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
			template.parts[i*2-1] = parseExpression(expression)
			template.parts[i*2].raw = subsplit[1]
		}
	}
	return template, nil
}

const (
	_          = iota
	SIMPLE     = iota
	PLUS       = iota
	SLASH      = iota
	DOT        = iota
	SEMICOLON  = iota
	QUERY      = iota
	AMPERSAND  = iota
	CROSSHATCH = iota
)

type templatePart struct {
	raw           string
	kind          int
	terms         []templateTerm
	first         string
	sep           string
	named         bool
	ifemp         string
	allowReserved bool
}

type templateTerm struct {
	name    string
	explode bool
}

func parseExpression(expression string) (result templatePart) {
	switch {
	case strings.HasPrefix(expression, "+"):
		result.kind = PLUS
		result.sep = ","
		result.allowReserved = true
	case strings.HasPrefix(expression, "."):
		result.kind = DOT
		result.first = "."
		result.sep = "."
	case strings.HasPrefix(expression, "/"):
		result.kind = SLASH
		result.first = "/"
		result.sep = "/"
	case strings.HasPrefix(expression, ";"):
		result.kind = SEMICOLON
		result.first = ";"
		result.sep = ";"
		result.named = true
	case strings.HasPrefix(expression, "?"):
		result.kind = QUERY
		result.first = "?"
		result.sep = "&"
		result.named = true
		result.ifemp = "="
	case strings.HasPrefix(expression, "&"):
		result.kind = AMPERSAND
		result.first = "&"
		result.sep = "&"
		result.named = true
		result.ifemp = "="
	case strings.HasPrefix(expression, "#"):
		result.kind = CROSSHATCH
		result.first = "#"
		result.sep = ","
		result.allowReserved = true
	default:
		result.kind = SIMPLE
		result.sep = ","
	}
	if result.kind != SIMPLE {
		expression = expression[1:]
	}
	rawterms := strings.Split(expression, ",")
	result.terms = make([]templateTerm, len(rawterms))
	for i, raw := range rawterms {
		result.terms[i] = parseTerm(raw)
	}
	return result
}

func parseTerm(term string) (result templateTerm) {
	if strings.HasSuffix(term, "*") {
		result.explode = true
		term = term[:len(term)-1]
	}
	split := strings.Split(term, ":")
	if len(split) == 1 {
		result.name = term
	} else if len(split) == 2 {
		result.name = split[0]
		// TODO: prefix modifier is in split[1]
	}
	// else error ?
	return result
}

func (self *UriTemplate) ExpandString(values map[string]interface{}) string {
	raw := ""
	for _, p := range self.parts {
		raw = raw + p.expand(values)
	}
	return raw
}

func (self *templatePart) expand(values map[string]interface{}) string {
	if len(self.raw) > 0 {
		return self.raw
	}
	result := self.first
	for _, term := range self.terms {
		value, exists := values[term.name]
		if !exists {
			continue
		}
		var next string
		switch value.(type) {
		case string:
			v := value.(string)
			next = self.expandString(term, v)
		case []interface{}:
			v := value.([]interface{})
			next = self.expandArray(term, v)
		case map[string]interface{}:
			v := value.(map[string]interface{})
			next = self.expandMap(term, v)
		default:
			continue
		}
		if result != self.first {
			result += self.sep
		}
		result += next
	}
	if result == self.first {
		result = ""
	}
	return result
}

func (self *templatePart) expandName(name string, empty bool) (result string) {
	if self.named {
		result = escape(name, self.allowReserved)
		if empty {
			result += self.ifemp
		} else {
			result += "="
		}
	}
	return result
}

func (self *templatePart) expandString(t templateTerm, s string) (result string) {
	return self.expandName(t.name, len(s) == 0) +
		escape(s, self.allowReserved)
}

func (self *templatePart) expandArray(t templateTerm, a []interface{}) (result string) {
	if !t.explode {
		result = self.expandName(t.name, len(a) == 0)
	}
	for i, v := range a {
		if t.explode && i > 0 {
			result += self.sep
		} else if i > 0 {
			result += ","
		}
		switch v.(type) {
		case string:
			s := v.(string)
			if self.named && t.explode {
				result += self.expandName(t.name, len(s) == 0)
			}
			result += escape(s, self.allowReserved)
		}
	}
	return result
}

func (self *templatePart) expandMap(t templateTerm, m map[string]interface{}) (result string) {
	for k, v := range m {
		if t.explode && len(result) > 0 {
			result += self.sep
		} else if len(result) > 0 {
			result += ","
		}
		switch v.(type) {
		case string:
			if t.explode {
				result += escape(k, self.allowReserved) +
					"=" + escape(v.(string), self.allowReserved)
			} else {
				result += escape(k, self.allowReserved) +
					"," + escape(v.(string), self.allowReserved)
			}
		default:
		}
	}
	if !t.explode {
		result = self.expandName(t.name, len(m) == 0) + result
	}
	return result
}
