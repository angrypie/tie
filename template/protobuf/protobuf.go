package protobuf

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	protogen "github.com/muxinc/protogen/proto3"
)

const moduleId = "protobuf"

func NewModule(p *parser.Parser) template.Module {
	return template.NewStandartModule(moduleId, Generate, p, nil)
}

func Generate(p *parser.Parser) (pkg *template.Package) {
	info := template.NewPackageInfoFromParser(p)
	//TODO all modules needs to create upgraded subpackage to make ServicePath reusable,

	spec := generateProtoSpec(info)

	fileStr, err := spec.Write()
	if err != nil {
		panic(err)
	}

	return &template.Package{
		Name:  moduleId,
		Files: [][]byte{[]byte(fileStr)},
	}
}

func generateProtoSpec(info *template.PackageInfo) (spec protogen.Spec) {
	spec.Package = info.PackageName
	template.ForEachFunction(info, true, func(fn parser.Function) {
		_, request, response := template.GetMethodTypes(fn)
		spec.Messages = append(spec.Messages, fieldsToMessage(request, fn.Arguments))
		spec.Messages = append(spec.Messages, fieldsToMessage(response, fn.Results.List()))
	})
	return spec
}

func fieldsToMessage(name string, fields []parser.Field) (message protogen.Message) {
	message.Name = name
	for _, field := range fields {
		message.Fields = append(message.Fields, protogen.CustomField{
			Name:   protogen.NameType(field.Name()),
			Typing: field.Type.TypeName(),
		})
	}

	return
}
