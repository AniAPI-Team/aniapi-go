package engine

import (
	"strings"
)

type Renderer struct {
	html     string
	response *Response
}

func (r *Renderer) ReplaceItem(k string, v string) {
	pattern := "<govar:" + k + " />"
	r.html = strings.Replace(r.html, pattern, v, -1)
}

func (r *Renderer) ReplaceList(k string, v []map[string]string) {
	pattern := "<gocycle:" + k + ">"
	from := strings.Index(r.html, pattern)

	if from == -1 {
		return
	}

	fromT := from + len(pattern)
	from -= 2

	pattern = "</gocycle:" + k + ">"
	toT := strings.Index(r.html, pattern)
	to := toT + len(pattern)
	toT -= 2

	template := strings.TrimSpace(r.html[fromT:toT])
	pattern = strings.TrimSpace(r.html[from:to])

	insertion := ""
	for _, m := range v {
		t := template

		for key, value := range m {
			t = strings.Replace(t, "<gocyclevar:"+key+" />", value, -1)
		}

		insertion += t
	}

	r.html = strings.Replace(r.html, pattern, insertion, 1)
}

func (r *Renderer) Serve() {
	r.response.Write(200, r.html)
}

func NewRenderer(html string, response *Response) *Renderer {
	return &Renderer{
		html:     html,
		response: response,
	}
}
