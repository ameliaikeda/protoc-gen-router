package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"

	"github.com/ameliaikeda/protoc-gen-router/service"
	"github.com/iancoleman/strcase"
	"github.com/monzo/terrors"
	"google.golang.org/protobuf/compiler/protogen"
)

const servicesTemplate = `package handler

import (
	"github.com/ameliaikeda/lib/proto"
	"github.com/monzo/typhon"

	{{ .Package }} "{{ .FullPackage }}"
)

func handle{{ .Name }}(req typhon.Request) typhon.Response {
	body := &{{ .Package }}.{{ .Request }}{}
	if err := proto.Decode(req, body); err != nil {
		return typhon.Response{Error: err}
	}

	// perform validation here.
	switch {
	case true:
		//
	}

	// do some database access here

	return proto.Encode(req, &{{ .Package }}.{{ .Response }}{})
}

`

var services = template.Must(template.New("protoc-gen-services").Parse(servicesTemplate))

func GenerateRPC(plugin *protogen.Plugin, svc *service.Service, rpc *service.RPC, folder string) error {
	filename := fmt.Sprintf("%s/%s.go", folder, strcase.ToSnake(rpc.Name))
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		// skip if we already have this file
		return nil
	}

	data := struct {
		Package     string
		FullPackage string
		Request     string
		Response    string
		Name        string
	}{
		Package:     svc.Package,
		FullPackage: svc.FullPackage,
		Request:     rpc.Request,
		Response:    rpc.Response,
		Name:        rpc.Name,
	}

	var buf bytes.Buffer
	if err := services.Execute(&buf, data); err != nil {
		return terrors.WrapWithCode(err, nil, "template.handler")
	}

	b, err := format.Source(buf.Bytes())
	if err != nil {
		return terrors.WrapWithCode(err, nil, "template.handler.format_source")
	}

	f := plugin.NewGeneratedFile(filename, protogen.GoImportPath(svc.FullPackage))

	written, err := f.Write(b)
	if err != nil {
		return terrors.WrapWithCode(err, nil, "template.router.write_file_failed")
	}

	if written != len(b) {
		return terrors.New("template.router.write_file_incomplete", fmt.Sprintf("Expected %d bytes written, got %d", len(b), written), nil)
	}

	return nil
}
