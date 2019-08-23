package template

import (
	"fmt"
	"log"
	"regexp"
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
	makeFirstNotEmptyStr(info, main, f)

}

func makeHTTPHandlers(info *PackageInfo, main *Group, f *File) {
	f.Comment(fmt.Sprintf("API handler methods (%s)", "HTTP")).Line()
	forEachFunction(info.Functions, func(fn *parser.Function) {
		makeHTTPHandler(info, fn, f.Group)
	})
}

func makeHTTPHandler(info *PackageInfo, fn *parser.Function, file *Group) {
	echo := "github.com/labstack/echo"
	handler, request, response := getMethodTypes(fn.Name, "HTTP")
	handlerBody := func(g *Group) {
		//Bind request params
		if len(fn.Arguments) > 0 {
			g.Id("request").Op(":=").New(Id(request))
			g.List(Id("_"), ListFunc(createArgsListFunc(fn.Arguments, "request", "string,"))).Op("=").
				List(Lit(0), ListFunc(createArgsList(fn.Arguments, func(arg *Statement) *Statement {
					return Id("firstNotEmptyStr").Call(
						Id("request").Dot(arg.GoString()),
						Id("ctx").Dot("QueryParam").Call(Lit(strings.ToLower(arg.GoString()))),
					)
				}, "", "string,")))
			g.If(
				Err().Op(":=").Id("ctx").Dot("Bind").Call(Id("request")),
				Err().Op("!=").Nil(),
			).Block(Return(Err()))
			//Empty argument needs to avoid errors if no other arguments exist
		}

		//Call original function
		g.Id("response").Op(":=").New(Id(response))

		//If method has receiver generate receiver middleware code
		//else just call public package method
		if receiver := fn.Receiver; receiver.Name != "" {
			receiverVar := "Receiver"
			g.Id(receiverVar).Op(":=").Qual(info.Service.Name, strings.Trim(receiver.Type, "*")).Block()
			makeReceiverMiddleware(receiverVar, g, findInitReceiver(info.Functions, fn))
			g.ListFunc(createArgsListFunc(fn.Results, "response")).
				Op("=").Id(receiverVar).Dot(fn.Name).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
		} else {
			g.ListFunc(createArgsListFunc(fn.Results, "response")).
				Op("=").Qual(info.Service.Name, fn.Name).Call(ListFunc(createArgsListFunc(fn.Arguments, "request")))
		}

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
	echoMiddleware := "github.com/labstack/echo/middleware"
	rndport := "github.com/angrypie/rndport"
	echo := "github.com/labstack/echo"

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

		forEachFunction(info.Functions, func(fn *parser.Function) {
			handler, _, _ := getMethodTypes(fn.Name, "HTTP")
			g.Id("server").Dot("POST").Call(
				Lit(strings.ToLower(fn.Name)),
				Id(handler),
			)
		})

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
					Return().List(Id("key").Op("==").Id("auth"), Nil()),
				)),
			)
		}
		g.Id("server").Dot("Start").Call(Id("address"))

	})
}

func makeFirstNotEmptyStr(info *PackageInfo, main *Group, f *File) {
	f.Func().Id("firstNotEmptyStr").Params(Id("a"), Id("b").String()).String().Block(
		If(Id("a").Op("!=").Lit("")).Block(Return(Id("a"))),
		Return(Id("b")),
	)
}

func forEachFunction(fns []*parser.Function, cb func(*parser.Function)) {
	for _, fn := range fns {
		//Skip InitReceiver middleware
		log.Println("process", fn.Name)
		if fn.Receiver.Name != "" && fn.Name == "InitReceiver" {
			log.Println("continue")
			continue
		}
		cb(fn)
	}
}

func findInitReceiver(fns []*parser.Function, forFunc *parser.Function) *parser.Function {
	for _, fn := range fns {
		if fn.Name != "InitReceiver" {
			continue
		}
		reg := fmt.Sprintf(`\A\*?%s\z`, forFunc.Receiver.Type)
		if match, _ := regexp.MatchString(reg, fn.Receiver.Type); !match {
			continue
		}

		return fn
	}
	return nil
}

func makeReceiverMiddleware(recId string, scope *Group, initReceiver *parser.Function) {
	if initReceiver == nil {
		return
	}
	scope.Id(recId).Dot("InitReceiver").CallFunc(func(g *Group) {
		for _, field := range initReceiver.Arguments {
			name := field.Name
			//TODO check function signature
			if name == "getHeader" {
				headerArg := "header"
				g.Func().Params(Id(headerArg).String()).String().Block(
					Return(Id("ctx").Dot("Request").Call().Dot("Header").Dot("Get").Call(Id(headerArg))),
				)
			}
		}
	})
}
