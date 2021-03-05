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

const ProtoMergeStyle = "protoMergeStyle"

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
func (m *MakeMergeModule) Execute(targets map[string]pgs.File, _ map[string]pgs.Package) []pgs.Artifact {

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
		return m.ListMerge(uccName)
	}
	if fld.Type().IsMap() {
		return m.MapMerge(uccName, fld)
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

func (m *MakeMergeModule) MapMerge(uccName pgs.Name, fld pgs.Field) string {
	base := `if x.%[1]s == nil || len(x.%[1]s) == 0 {
						x.%[1]s = d.%[1]s
					}`
	if _, ok := m.ctx.Params()[ProtoMergeStyle]; ok {
		protoStyleMerge := ` else {
						for k, v := range d.%[1]s {
							if _, present := x.%[1]s[k]; !present {
								x.%[1]s[k] = v
							} %s
						}
					}`
		if fld.Type().Element().IsEmbed() {
			return fmt.Sprintf(base+protoStyleMerge, uccName,
				fmt.Sprintf(`else {
						x.%[1]s[k].Merge(v)
				}`, uccName))
		}
		return fmt.Sprintf(base+protoStyleMerge, uccName, "")
	}
	return fmt.Sprintf(base, uccName)
}

func (m *MakeMergeModule) ListMerge(uccName pgs.Name) string {
	base := `if x.%[1]s == nil || len(x.%[1]s) == 0 {
						x.%[1]s = d.%[1]s
					}`
	if _, ok := m.ctx.Params()[ProtoMergeStyle]; ok {
		//It's not a set, I will NOT equality-check to prevent duplicates
		return fmt.Sprintf(base+` else {
						x.%[1]s = append(x.%[1]s, d.%[1]s...)
						}`, uccName)
	}
	return fmt.Sprintf(base, uccName)
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
