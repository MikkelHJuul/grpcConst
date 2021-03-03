package main

import (
	"fmt"
	"text/template"

	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

type MakeMergeModule struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
	tpl *template.Template
}

func MakeMerge() *MakeMergeModule {
	return &MakeMergeModule{ModuleBase: &pgs.ModuleBase{}}
}

func (m *MakeMergeModule) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.ctx = pgsgo.InitContext(c.Parameters())

	tpl := template.New(m.Name()).Funcs(map[string]interface{}{
		"package":    m.ctx.PackageName,
		"name":       m.ctx.Name,
		"writeField": m.writeField,
	})

	m.tpl = template.Must(tpl.Parse(mergeTpl))
}

func (m *MakeMergeModule) Name() string {
	return "mergeFunctions"
}
func (m *MakeMergeModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {

	for _, t := range targets {
		m.generate(t)
	}

	return m.Artifacts()
}

func (m *MakeMergeModule) generate(f pgs.File) {
	if len(f.Messages()) == 0 {
		return
	}

	name := m.ctx.OutputPath(f).SetExt(".merge.go")
	m.AddGeneratorTemplateFile(name.String(), m.tpl, f)
}

func (m *MakeMergeModule) writeField(fld pgs.Field) string {
	if fld.InOneOf() {
		return fmt.Sprintf("//OneOf field -- %s -- not touching this atm.", fld.Name())
	}
	uccName := pgsgo.PGGUpperCamelCase(fld.Name())
	if fld.Type().IsRepeated() {
		return fmt.Sprintf(
			`if x.%[1]s == nil || len(x.%[1]s) == 0 {
						x.%[1]s = d.%[1]s
					}`, uccName)
	}
	switch fld.Type().ProtoType() {
	case pgs.Int64T, pgs.UInt64T, pgs.SFixed64, pgs.SInt64, pgs.Fixed64T,
		pgs.Int32T, pgs.UInt32T, pgs.SFixed32, pgs.SInt32, pgs.Fixed32T, pgs.DoubleT, pgs.FloatT: // isNumeric
		return fmt.Sprintf(
			`if x.%[1]s == 0 {
    					x.%[1]s = d.%[1]s
					}`, uccName)
	case pgs.StringT:
		return fmt.Sprintf(
			`if x.%[1]s == "" {
    					x.%[1]s = d.%[1]s
					}`, uccName)
	case pgs.MessageT:
		return fmt.Sprintf(
			`if x.%[1]s == nil {
						x.%[1]s = d.%[1]s
					} else {
						x.%[1]s.Merge(d.%[1]s)
					}`, uccName)
	default: // pgs.BoolT, pgs.EnumT, pgs.GroupT:
		return fmt.Sprintf(`// fallthrough type: %s`, fld.Type().ProtoType().String())
	}

}

const mergeTpl = `package {{ package . }}

{{ range .AllMessages }}

func (x *{{ name . }}) Merge(donor interface{}) {
	if d, ok := donor.({{ name . }}); ok {
	{{ range .Fields }}
		{{ writeField . }}
	{{ end }}
	}
}

{{ end }}
`
