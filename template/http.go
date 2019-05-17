package template

import (
	"fmt"
	"strings"

	"github.com/angrypie/tie/parser"
	. "github.com/dave/jennifer/jen"
)

func makeHTTPServer(info *PackageInfo, main *Group, f *File) {
	service := info.Service
	if service.Type != "http" && service.Type != "httpOnly" {
		return
	}
	makeStartHTTPServer(info, main, f)
	makeHTTPRequestResponseTypes(info, main, f)
	makeHTTPHandlers(info, main, f)

}

func makeHTTPHandlers(info *PackageInfo, main *Group, f *File) {
	f.Comment(fmt.Sprintf("API handler methods (%s)", "HTTP")).Line()
	for _, fn := range info.Functions {
		makeHTTPHandler(info, fn, f.Group)
	}
}

func makeHTTPHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	echo := "github.com/labstack/echo"
	handler, request, response := getMethodTypes(fn.Name, "HTTP")
	handlerBody := func(g *Group) {
		//Bind request params
		if len(fn.Arguments) > 0 {
			g.Id("request").Op(":=").New(Id(request))
			g.If(
				Err().Op(":=").Id("ctx").Dot("Bind").Call(Id("request")),
				Err().Op("!=").Nil(),
			).Block(Return(Err()))
		}

		//Call original function
		g.Id("response").Op(":=").New(Id(response))
		g.ListFunc(createArgsListFunc(fn.Results, "response")).
			Op("=").Qual(info.Service.Name, fn.Name).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))

		g.If(Err().Op(":=").Id("response").Dot("Err"), Err().Op("!=").Nil()).Block(
			Return(Id("ctx").Dot("JSON").Call(
				Qual("net/http", "StatusBadRequest"),
				Map(String()).String().Values(Dict{Lit("err"): Err().Dot("Error").Call()}),
			)),
		)

		g.Return(Id("ctx").Dot("JSON").Call(Qual("net/http", "StatusOK"), Id("response")))
	}

	file.Func().Id(handler).
		Params(Id("ctx").Qual(echo, "Context")).
		Params(Err().Error()).BlockFunc(handlerBody).
		Line()
}

func makeHTTPRequestResponseTypes(info *PackageInfo, main *Group, f *File) {
	f.Add(createReqRespTypes("HTTP", info))
}

func makeStartHTTPServer(info *PackageInfo, main *Group, f *File) {
	echo := "github.com/labstack/echo"
	echoMiddleware := "github.com/labstack/echo/middleware"
	rndport := "github.com/angrypie/rndport"

	main.Go().Id("startHTTPServer").Call()

	f.Func().Id("startHTTPServer").Params().BlockFunc(func(g *Group) {
		port := info.Service.Port
		if port == "" {
			g.List(Id("address"), Err()).Op(":=").
				Qual(rndport, "GetAddress").Call(Lit(":%d"))
			g.If(Err().Op("!=").Nil()).Block(Panic(Err()))
		} else {
			g.Id("address").Op(":=").Lit(fmt.Sprintf(":%s", port))
		}

		g.Id("server").Op(":=").Qual(echo, "New").Call()
		g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "CORSWithConfig").Call(
			Qual(echoMiddleware, "CORSConfig").Values(Dict{
				Id("AllowOrigins"): Index().String().Values(Lit("*")),
			}),
		))

		for _, fn := range info.Functions {
			handler, _, _ := getMethodTypes(fn.Name, "HTTP")
			g.Id("server").Dot("POST").Call(
				Lit(strings.ToLower(fn.Name)),
				Id(handler),
			)
		}

		if key := info.Service.Auth; key != "" {
			g.Id("server").Dot("Use").Call(Qual(echoMiddleware, "KeyAuth").Call(
				Func().Params(Id("key").String(), Id("ctx").Qual(echo, "Context")).Params(Bool(), Error()).Block(
					Id("auth").Op(":=").Lit(key),
					If(
						Id("envKey").Op(":=").Qual("os", "Getenv").Call(Lit("TIE_API_KEY")),
						Id("envKey").Op("!=").Lit(""),
					).Block(
						Id("auth").Op("=").Id("envKey"),
					),
					Return().List(Id("key").Op("==").Id("auth"), Error()),
				)),
			)
		}
		g.Id("server").Dot("Start").Call(Id("address"))

	})
}
