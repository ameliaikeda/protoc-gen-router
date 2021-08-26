package main

import (
	"github.com/monzo/terrors"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/ameliaikeda/protoc-gen-router/generator"
	"github.com/ameliaikeda/protoc-gen-router/service"
)

func main() {
	opts := &protogen.Options{}
	opts.Run(func(plugin *protogen.Plugin) error {
		if len(plugin.Request.FileToGenerate) != 1 {
			return terrors.BadRequest("protobuf", "Router generation is only valid with a single input proto file.", nil)
		}

		for _, proto := range plugin.Request.ProtoFile {
			// we will only code-gen if a single service is present.
			if len(proto.Service) != 1 {
				continue
			}

			svc := proto.Service[0]

			// if there are no methods on the service, we will not code-gen.
			if len(svc.Method) == 0 {
				continue
			}

			// grab a "service" intermediate type
			s, err := service.New(proto, svc)
			if err != nil {
				return err
			}

			if err := generator.GenerateRouter(plugin, s); err != nil {
				return err
			}

			// now we have the file, the request and the service, pass it to each generator method
			if err := generator.GenerateHandler(plugin, s); err != nil {
				return err
			}
		}

		return nil
	})
}
