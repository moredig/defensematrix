package security

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	coreruleset "github.com/corazawaf/coraza-coreruleset/v4"
	"github.com/corazawaf/coraza/v3"
	corazahttp "github.com/corazawaf/coraza/v3/http"
)

const requestBodyLimit = 1 << 20

const directives = `
Include @coraza.conf-recommended
SecRuleEngine On
SecRequestBodyAccess On
SecRequestBodyLimit 1048576
SecRequestBodyNoFilesLimit 1048576
SecRequestBodyLimitAction Reject
SecResponseBodyAccess Off
Include @crs-setup.conf.example
Include @owasp_crs/*.conf
`

type portableFS struct {
	fs.FS
}

func (p portableFS) Open(name string) (fs.File, error) {
	return p.FS.Open(filepathName(name))
}

func (p portableFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(p.FS, filepathName(name))
}

func (p portableFS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(p.FS, filepathName(name))
}

func (p portableFS) Glob(pattern string) ([]string, error) {
	return fs.Glob(p.FS, filepathName(pattern))
}

func filepathName(name string) string {
	return path.Clean(strings.ReplaceAll(name, "\\", "/"))
}

// WrapHTTP applies the OWASP Core Rule Set before forwarding HTTP requests.
func WrapHTTP(next http.Handler) (http.Handler, error) {
	waf, err := coraza.NewWAF(coraza.NewWAFConfig().
		WithDirectives(directives).
		WithRootFS(portableFS{FS: coreruleset.FS}))
	if err != nil {
		return nil, err
	}
	return corazahttp.WrapHandler(waf, next), nil
}
