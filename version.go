package formation

import (
	"fmt"
	"runtime"
	"strings"
	"text/template"

	"xsky.com/sds-formation/autogen/version"
)

// Version returns current version
func Version() string {
	return fmt.Sprintf("formation version: %s, build %s",
		version.VERSION, version.GITCOMMIT)
}

var (
	detailedVersionTmplStr = `{{.App}} Version: {{.Version}}
Git SHA: {{.GitCommit}}
Go Version: {{.GoVersion}}
Go OS/Arch: {{.OS}}/{{.Arch}}`

	detailedVersionTmpl = template.Must(
		template.New("DetailedVersion").Parse(detailedVersionTmplStr))
)

// DetailedVersion returns Go and software version
func DetailedVersion() string {
	v := &struct {
		App       string
		Version   string
		GitCommit string
		GoVersion string
		OS        string
		Arch      string
	}{
		App:       "xsky.com/sds-formation",
		Version:   version.VERSION,
		GitCommit: version.GITCOMMIT,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
	b := &strings.Builder{}
	detailedVersionTmpl.Execute(b, v)
	return b.String()
}
