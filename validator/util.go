package validator

import (
	"github.com/prometheus/alertmanager/template"
	"io/ioutil"
	"os"
)

// TemplateForTestsString is the template used for unit tests and integration tests.
// We have it separate from above default template because any tiny change in the template
// will require updating almost all channel tests (15+ files) and it's very time consuming.
const TemplateForTestsString = `
{{/* begin my title template mod from default.title */}}
{{ define "my.title" }} [{{ .Alerts.Firing | len }} firing, {{ .Alerts.Resolved | len }} resolved] {{ (index .Alerts 0).Annotations.summary }}{{ end }}
{{/* end my.title */}}

{{/* begin define my.message template mod from default.message */}}
{{ define "my__text_alert_list" }}{{ range . }}

Annotations:
{{ range .Annotations.SortedPairs }} - {{ .Name }} = {{ .Value }}
{{ end }}

Labels:
{{ range .Labels.SortedPairs }} - {{ .Name }} = {{ .Value }}
{{ end }}

{{/* only show DashboardURL if no PanelURL */}}
{{ if and (gt (len .DashboardURL) 0) (eq (len .PanelURL) 0) }}Dashboard: {{ .DashboardURL }}
{{ end }}

{{ if gt (len .PanelURL) 0 }}Panel: {{ .PanelURL }}
{{ end }}

{{ end }} {{/* end range */}}

{{ end }} {{/* end define __text_alert_list */}}

{{/* my message template */}}
{{ define "default.message" }}

{{ if gt (len .Alerts.Firing) 0 }} 🔥 *Firing*
{{ template "my__text_alert_list" .Alerts.Firing }}
{{ end }}

{{ if gt (len .Alerts.Resolved) 0 }} ✅ *Resolved*
{{ template "my__text_alert_list" .Alerts.Resolved }}
{{ end }}

{{ end }}
{{/* end define my.message */}}
`

func ErrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// TemplateForTests write template string to tmp files for test purposes
func TemplateForTests(tmplstr string) *template.Template {
	f, err := ioutil.TempFile("/tmp", "template")
	ErrPanic(err)

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	defer func() {
		err := os.RemoveAll(f.Name())
		ErrPanic(err)
	}()

	_, err = f.WriteString(tmplstr)
	ErrPanic(err)

	tmpl, err := template.FromGlobs(f.Name())
	ErrPanic(err)

	return tmpl
}
