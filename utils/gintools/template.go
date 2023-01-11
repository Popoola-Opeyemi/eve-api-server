package gintools

// cSpell:ignore mkr, gocraft, gommon, Sprintf, dbname, Infof, dannyvankooten, extemplate
import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"

	"eve/utils"

	"github.com/dannyvankooten/extemplate"
	"github.com/gin-gonic/gin/render"
	"go.uber.org/zap"
)

// TemplateDefinition ...
type TemplateDefinition struct {
	Name string
	Glob string
}

// TemplateMgr ...
type TemplateMgr struct {
	path string
	tpl  *extemplate.Extemplate
	defs []TemplateDefinition
	log  *zap.SugaredLogger
	cfg  *utils.Config
}

// NewTemplateMgr creates a new instance of TemplateMgr
func NewTemplateMgr(cfg *utils.Config, log *zap.SugaredLogger, path string) (*TemplateMgr, error) {
	mgr := &TemplateMgr{}

	mgr.path = path
	mgr.log = log

	if err := mgr.Init(); err != nil {
		return nil, err
	}

	return mgr, nil
}

// Init ...
func (s *TemplateMgr) Init() (err error) {

	s.tpl = extemplate.New()
	s.tpl.Funcs(template.FuncMap{
		"json": marshalJSON,
		"mix":  s.mixAsset,
	})
	err = s.tpl.ParseDir(s.path, []string{".html"})

	return
}

// Instance ...
func (s *TemplateMgr) Instance(name string, data interface{}) render.Render {
	tpl := s.tpl.Lookup(name)
	if tpl != nil {
		html := render.HTML{
			Template: tpl,
			Name:     name,
			Data:     data,
		}

		return html
	}
	return nil
}

// Render ...
func (s *TemplateMgr) Render(w io.Writer, name string, data interface{}) error {
	if err := s.tpl.ExecuteTemplate(w, name, data); err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func marshalJSON(val interface{}) template.JS {
	retv := []byte{}
	retv, _ = json.Marshal(val)

	return template.JS(retv)
}

// mixAsset reads a laravel-mix mix-manifest.json file
// and returns the hashed filename.
// assumes that the file will be in ./static
func (s TemplateMgr) mixAsset(val string) string {

	manifest := filepath.Join(s.cfg.FullPath("static"), "mix-manifest.json")
	content, err := ioutil.ReadFile(manifest)
	if err != nil {
		s.log.Error(err)
		return fmt.Sprintf("err cant read mix-manifest")
	}

	data := map[string]string{}
	if err := json.Unmarshal(content, &data); err != nil {
		s.log.Error(err)
		return fmt.Sprintf("err cant unmarshal mix-manifest")
	}

	retv, found := data[val]
	if !found {
		return fmt.Sprintf("err cant find %s mix-manifest", val)
	}

	return s.cfg.FullURL("static") + retv
}
