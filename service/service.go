package service

import (
	"fmt"
	"path"
	"strings"

	routerproto "github.com/ameliaikeda/protoc-gen-router/proto"
	"github.com/monzo/terrors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Service is a representation of a service proto.
type Service struct {
	// Filename is the name of the file we're generating.
	Filename string

	// Package is the go package name.
	Package string

	// FullPackage is the entire go import path via option go_package
	FullPackage string

	// Name is the name of this service verbatim, snake_case
	Name string

	// Host is the name of this service, as a fully qualified host (service.foo-bar)
	Host string

	// ServiceType is the "type" we specified, e.g. "service", "api", "graphql"
	ServiceType string

	// RPCs is a list of all RPCs in this service, with options attached to discover the path.
	RPCs []*RPC

	// Directory is the absolute directory that the service is contained within, without a trailing '/'.
	Directory string
}

// RPC is a field representing an RPC call in a service.
type RPC struct {
	// Method is the HTTP method for this RPC call, derived from the Request name.
	Method string

	// Path is the path to this RPC, starting with /
	Path string

	// Request is the Go type of the request.
	Request string

	// Response is the Go type of the response.
	Response string

	// Future is the "Request" name, with the suffix replaced with Future.
	// It is a Go type.
	Future string

	// Name is the name of the RPC on a service.
	Name string
}

func New(file *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto) (*Service, error) {
	service := &Service{
		Filename: file.GetName(),
		Package:  file.GetPackage(),
		Name:     svc.GetName(),
		RPCs:     make([]*RPC, 0, len(svc.Method)),
	}

	fileOpts := file.GetOptions()
	if fileOpts == nil {
		return nil, terrors.New("protobuf.options.go_package", "failed to get go_package option", nil)
	}

	service.FullPackage = strings.Split(fileOpts.GetGoPackage(), ";")[0]
	service.Directory = strings.Replace(service.Filename, "/proto/"+path.Base(service.Filename), "", 1)

	if svc.Options == nil {
		return nil, terrors.New("protobuf.options", "failed to get router options - did you add the service name?", nil)
	}

	opts := proto.GetExtension(svc.Options, routerproto.E_Router)
	if opts == nil {
		return nil, terrors.New("protobuf.extension.router", "failed to get router extension", nil)
	}

	if o, ok := opts.(*routerproto.ServiceRouter); ok && o != nil {
		service.Host = o.Name
		service.ServiceType = o.ServiceType
	}

	for _, rpc := range svc.Method {
		// get options for method.
		if rpc.Options == nil {
			return nil, terrors.New("protobuf.rpc.options", "Unable to find path. Have you used option (handler).path?", nil)
		}

		opts := proto.GetExtension(rpc.Options, routerproto.E_Handler)
		if opts == nil {
			return nil, terrors.New("protobuf.extension.handler", "failed to get handler extension", nil)
		}

		r, ok := opts.(*routerproto.RPCHandler)
		if !ok || r == nil {
			return nil, terrors.New("protobuf.rpc.handler", "option (handler) is not an RPC handler. Did you import the right .proto file?", nil)
		}

		service.RPCs = append(service.RPCs, rpcFromProto(file.GetPackage(), r, rpc))
	}

	return service, nil
}

func rpcFromProto(pkg string, opts *routerproto.RPCHandler, rpc *descriptorpb.MethodDescriptorProto) *RPC {
	prefix := fmt.Sprintf(".%s.", pkg)

	request := strings.TrimPrefix(rpc.GetInputType(), prefix)
	response := strings.TrimPrefix(rpc.GetOutputType(), prefix)
	name := strings.TrimPrefix(rpc.GetName(), prefix)
	future := fmt.Sprintf("%sFuture", strings.TrimSuffix(request, "Request"))
	method := getMethod(opts, request)

	return &RPC{
		Path:     opts.Path,
		Method:   method,
		Request:  request,
		Response: response,
		Future:   future,
		Name:     name,
	}
}

func getMethod(opts *routerproto.RPCHandler, name string) string {
	if opts.Method != "" {
		return strings.ToUpper(opts.Method)
	}

	methods := [...]string{
		"GET",
		"PUT",
		"POST",
		"PATCH",
		"DELETE",
	}

	for _, method := range methods {
		if strings.HasPrefix(name, method) {
			return method
		}
	}

	return ""
}
