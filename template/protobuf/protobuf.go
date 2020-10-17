package protobuf

import (
	"hash/fnv"
	"log"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
	"github.com/angrypie/tie/template/modutils"
	protogen "github.com/muxinc/protogen/proto3"
)

const moduleId = "protobuf"

func NewModule(p *parser.Parser) modutils.Module {
	return modutils.NewStandartModule(moduleId, Generate, p, nil)
}

func Generate(p *parser.Parser) (pkg *modutils.Package) {
	info := template.NewPackageInfoFromParser(p)
	//TODO all modules needs to create upgraded subpackage to make ServicePath reusable,

	spec := generateProtoSpec(info)

	fileStr, err := spec.Write()
	if err != nil {
		panic(err)
	}

	return modutils.NewPackage(moduleId, "schema.proto", fileStr)

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
		fieldType := field.Type.TypeName()
		if fieldType == "error" {
			fieldType = "string"
		}

		//TODO provide prod ready solution to generate field numbers
		log.Println("TODO: protobuf fields generation is not production ready")
		fieldNumber := hashFieldNameToNumber(field.Name())
		message.Fields = append(message.Fields, protogen.CustomField{
			Name:   protogen.NameType(field.Name()),
			Typing: field.Type.TypeName(),
			Tag:    protogen.TagType(fieldNumber),
		})
	}

	return
}

//hashFieldNameToNumber is tmp solution for protobuf field numbers genraation.
func hashFieldNameToNumber(s string) uint8 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return uint8(h.Sum32() % 2047)
}
