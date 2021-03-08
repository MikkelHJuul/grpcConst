package main

import (
	"fmt"
	"github.com/envoyproxy/protoc-gen-validate/module"
	"strings"
	"text/template"

	"github.com/MikkelHJuul/grpcConst/cmd/protoc-gen-merge/merge"

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

//Match is almost a Copy paste of the GoFmt PostProcessor from PG* (with other SUFFIX)
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

	return strings.HasSuffix(n, "reduce.go.txt")
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

	name := r.ctx.OutputPath(f).SetExt(".reduce.go.txt")
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
	return r.writeFieldName(fld, string("x."+uccName), string("r."+uccName))
}

func (r *MakeReduceModule) writeFieldName(fld interface{}, rcv, don string) string {
	var prototype pgs.ProtoType
	if fldType, ok := fld.(module.FieldType); ok {
		prototype = fldType.ProtoType()
	}
	if pgsField, ok := fld.(pgs.Field); ok {
		pgsType := pgsField.Type()
		if pgsType.IsRepeated() {
			notEqual := rcv + "[i] != " + don + "[i]"
			if pgsType.Element().IsEmbed() {
				//TODO - EQUALITY
				notEqual = "!reflect.DeepEquals(" + rcv + "[i], " + don + "[i])"
			}
			return fmt.Sprintf(
				`if %[1]s != nil && %[2]s != nil && len(%[1]s) == len(%[2]s) {
						shouldRemove = true
						i := 0
						for shouldRemove {
							if `+notEqual+` {
								shouldRemove = false
							}
							i++
						}
						if shouldRemove {
							%[1]s = nil
						}
					}`, rcv, don)
		}
		if pgsType.IsMap() {
			return r.reduceMap(rcv, don, pgsField)
		}
		prototype = pgsType.ProtoType()
	}
	switch prototype {
	case pgs.Int64T, pgs.UInt64T, pgs.SFixed64, pgs.SInt64, pgs.Fixed64T,
		pgs.Int32T, pgs.UInt32T, pgs.SFixed32, pgs.SInt32, pgs.Fixed32T, pgs.DoubleT, pgs.FloatT: // isNumeric
		return fmt.Sprintf(
			`if %[1]s == %[2]s {
    					%[1]s = 0
					}`, rcv, don)
	case pgs.StringT:
		return fmt.Sprintf(
			`if %[1]s == %[2]s {
    					%[1]s = ""
					}`, rcv, don)
	case pgs.BytesT:
		return fmt.Sprintf(
			`if bytes.Equal(%[1]s, %[2]s) {
    					x.%[1]s = nil
					}`, rcv, don)
	case pgs.MessageT:
		return fmt.Sprintf(
			`if %[1]s != nil {
						%[1]s.Reduce(%[1]s)
					}`, rcv, don)
	default: // pgs.BoolT, pgs.EnumT, pgs.GroupT
		r.Logf("Warning, your compiled code contains code that cannot be reduced: %s", prototype.String())
		return fmt.Sprintf(`// fallthrough type: %s`, prototype.String())
	}
}

func (r *MakeReduceModule) reduceMap(rcv, don string, fld pgs.Field) (mapReduction string) {
	notEquals := "rv != v"
	if fld.Type().Element().IsEmbed() {
		notEquals = "!reflect.DeepEquals(rv, v)" //could recurse Reduce.. but it's not compatible with reflection/merge implementation.
	}
	mapReduction = fmt.Sprintf(
		`if %[1]s != nil && %[2]s != nil && len(%[1]s) == len(%[2]s) {
						shouldRemove := true
						for k, v := range %[1]s {
							if rv, ok := %[2]s[k]; !ok || `+notEquals+` {
								shouldRemove = false
								break;
							}
						}
						if shouldRemove {
							%[1]s = nil
						}
					}`, rcv, don)
	if _, ok := r.ctx.Params()[merge.ProtoMergeStyle]; ok {
		mapReduction = fmt.Sprintf(
			`for k, v := range %[1]s {
						if rv, ok := %[2]s[k]; ok {
							`+r.writeFieldName(fld.Type().Element(), "rv", "v")+`
						}
					}`, don, rcv)
	}
	return
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
