package main

import (
	"fmt"
	"strings"
	"text/template"

	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

type MakeReduceModule struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
	tpl *template.Template
}

func MakeReduce() *MakeReduceModule {
	return &MakeReduceModule{ModuleBase: &pgs.ModuleBase{}}
}

type ReducePostProcessor struct{}

func (p ReducePostProcessor) Match(a pgs.Artifact) bool {
	var n string

	switch a := a.(type) {
	case pgs.GeneratorFile:
		n = a.Name
	case pgs.GeneratorTemplateFile:
		n = a.Name
	case pgs.CustomFile:
		n = a.Name
	case pgs.CustomTemplateFile:
		n = a.Name
	default:
		return false
	}

	return strings.HasSuffix(n, "reduce.go")
}

func (p ReducePostProcessor) Process(in []byte) ([]byte, error) {
	asString := string(in)
	var imports []string
	if strings.Contains(asString, "bytes.Equal") {
		imports = []string{"bytes"}
	}
	if strings.Contains(asString, "reflect.DeepEquals") {
		imports = append(imports, "reflect")
	}
	importString := generateImportString(imports)
	asString = strings.Replace(asString, importsStatement, importString, 1)
	return []byte(asString), nil
}

func generateImportString(imports []string) string {
	switch len(imports) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf(`import "%s"`, imports[0])
	default:
		return fmt.Sprintf(`import (
    %s
)`, fmt.Sprintf(`"%s"`, strings.Join(imports, `"
"`)))
	}

}

func AddImports() pgs.PostProcessor {
	return ReducePostProcessor{}
}

func (r *MakeReduceModule) InitContext(c pgs.BuildContext) {
	r.ModuleBase.InitContext(c)
	r.ctx = pgsgo.InitContext(c.Parameters())

	tpl := template.New(r.Name()).Funcs(map[string]interface{}{
		"package":    r.ctx.PackageName,
		"name":       r.ctx.Name,
		"writeField": r.writeField,
	})

	r.tpl = template.Must(tpl.Parse(reduceTpl))
}

func (r *MakeReduceModule) Name() string {
	return "reduceFunctions"
}

func (r *MakeReduceModule) Execute(targets map[string]pgs.File, _ map[string]pgs.Package) []pgs.Artifact {

	for _, t := range targets {
		r.generate(t)
	}

	return r.Artifacts()
}

func (r *MakeReduceModule) generate(f pgs.File) {
	if len(f.Messages()) == 0 {
		return
	}

	name := r.ctx.OutputPath(f).SetExt(".reduce.go")
	r.AddGeneratorTemplateFile(name.String(), r.tpl, f)

	someBool := true
	i := 0

	for someBool {
		if i == 10 {
			someBool = false
		}
		i++
	}
}

func (r *MakeReduceModule) writeField(fld pgs.Field) string {
	if fld.InOneOf() {
		return fmt.Sprintf("//OneOf field -- %s -- not touching this atm.", fld.Name())
	}
	uccName := pgsgo.PGGUpperCamelCase(fld.Name())
	if fld.Type().IsRepeated() {
		return fmt.Sprintf(
			`if x.%[1]s != nil && r.%[1]s != nil && len(x.%[1]s) == len(r.%[1]s) {
						shouldRemove = true
						i := 0
						for shouldRemove {
							if x.%[1]s[i] != r.%[1]s[i] {
								shouldRemove = false
							}
							i++
						}
						if shouldRemove {
							r.%[1]s = nil
						}
					}`, uccName)
	}
	if fld.Type().IsMap() {
		return fmt.Sprintf(
			`if x.%[1]s != nil && r.%[1]s != nil && len(x.%[1]s) == len(r.%[1]s) {
						shouldRemove = true
						for k, v := range x.%[1]s {
							if rv, ok := r.%[1]s[k]; !ok || reflect.DeepEquals(v, rv) {
								shouldRemove = false
								break;
							}
						}
						if shouldRemove {
							r.%[1]s = nil
						}
					}`, uccName)
	}
	switch fld.Type().ProtoType() {
	case pgs.Int64T, pgs.UInt64T, pgs.SFixed64, pgs.SInt64, pgs.Fixed64T,
		pgs.Int32T, pgs.UInt32T, pgs.SFixed32, pgs.SInt32, pgs.Fixed32T, pgs.DoubleT, pgs.FloatT: // isNumeric
		return fmt.Sprintf(
			`if x.%[1]s == r.%[1]s {
    					x.%[1]s = 0
					}`, uccName)
	case pgs.StringT:
		return fmt.Sprintf(
			`if x.%[1]s == r.%[1]s {
    					x.%[1]s = ""
					}`, uccName)
	case pgs.BytesT:
		return fmt.Sprintf(
			`if bytes.Equal(x.%[1]s, r.%[1]s) {
    					x.%[1]s = nil
					}`, uccName)
	case pgs.MessageT:
		return fmt.Sprintf(
			`if x.%[1]s != nil {
						x.%[1]s.Reduce(r.%[1]s)
					}`, uccName)
	default: // pgs.BoolT, pgs.EnumT, pgs.GroupT
		r.Logf("Warning, your compiled code contains code that cannot be reduced: %s", fld.FullyQualifiedName())
		return fmt.Sprintf(`// fallthrough type: %s`, fld.Type().ProtoType().String())
	}
}

const importsStatement = `//imports here`
const reduceTpl = `package {{ package . }}
//This code is generated and should not be edited

` + importsStatement + `

{{ range .AllMessages }}

func (x *{{ name . }}) Reduce(reference interface{}) {
	if r, ok := reference.({{ name . }}); ok {
	{{ range .Fields }}
		{{ writeField . }}
	{{ end }}
	}
}

{{ end }}
`
