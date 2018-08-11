package helm

import (
	"fmt"
	"regexp"
	"strings"

	"yunion.io/x/log"
)

const (
	IngressKind           = "Ingress"
	PodSecurityPolicyKind = "PodSecurityPolicy"
)

var (
	resSplit  = regexp.MustCompile(`\s{2,}`)
	attrSplit = regexp.MustCompile(`\s+`)

	ResourceEmptyAttrsMap = map[string][]string{
		IngressKind:           {"ADDRESS"},
		PodSecurityPolicyKind: {"SUPGROUP"},
	}
)

type Resource map[string]string

type Resources []Resource

type resourceParser struct {
	kind     string
	attrs    []string
	contents []string
}

func (r *resourceParser) removeEmptyAttrs(values []string) error {
	removeAttr := func(attrs []string, key string) []string {
		ret := make([]string, 0)
		for _, attr := range attrs {
			if attr == key {
				continue
			}
			ret = append(ret, attr)
		}
		return ret
	}
	emptyAttr, ok := ResourceEmptyAttrsMap[r.kind]
	if !ok {
		return fmt.Errorf("Not found %q kind resource empty attrs", r.kind)
	}
	for _, v := range emptyAttr {
		if len(values) == len(r.attrs) {
			break
		}
		r.attrs = removeAttr(r.attrs, v)
	}
	if len(r.attrs) != len(values) {
		return fmt.Errorf("kind: %s, parser attrs: %#v, values: %#v not match", r.kind, r.attrs, values)
	}
	return nil
}

func (r *resourceParser) toResource(content string) (ret Resource, err error) {
	attrsV := attrSplit.Split(content, len(r.attrs))
	if len(r.attrs) != len(attrsV) {
		if len(r.attrs) > len(attrsV) {
			err = r.removeEmptyAttrs(attrsV)
			if err != nil {
				err = fmt.Errorf("RemoveEmptyAttrs: %v", err)
				return
			}
		} else {
			err = fmt.Errorf("kind: %s, attrs: %#v less than values: %#v", r.kind, r.attrs, attrsV)
			return
		}
	}
	ret = make(map[string]string)
	ret["kind"] = r.kind
	for i, key := range r.attrs {
		ret[strings.ToLower(key)] = attrsV[i]
	}
	return
}

func (r *resourceParser) Resources() Resources {
	ret := make([]Resource, 0)
	for _, content := range r.contents {
		res, err := r.toResource(content)
		if err != nil {
			log.Errorf("Parse content: %q, err: %v", content, err)
			continue
		}
		ret = append(ret, res)
	}
	return ret
}

func newResourceParser(line string) *resourceParser {
	lines := strings.Split(line, "\n")
	if len(lines) <= 2 {
		return nil
	}
	if !strings.Contains(lines[0], "==>") {
		return nil
	}
	vkind := strings.Split(lines[0], "/")
	if len(vkind) < 2 {
		log.Errorf("Invalid kind version line: %q", lines[0])
		return nil
	}
	kind := vkind[1]
	attrs := resSplit.Split(lines[1], -1)
	contents := lines[2:]
	return &resourceParser{
		kind:     kind,
		attrs:    attrs,
		contents: contents,
	}
}

func resourcesLines(str string) []string {
	return strings.Split(str, "\n\n")
}

func ParseResources(resStr string) (ret map[string]Resources) {
	ret = make(map[string]Resources, 0)
	for _, line := range resourcesLines(resStr) {
		if len(line) == 0 {
			continue
		}
		p := newResourceParser(line)
		if p == nil {
			log.Errorf("Invalid line: %q", line)
			continue
		}
		ret[p.kind] = p.Resources()
	}
	return
}
