# protoc-gen-router

protoc-gen-router generates routers from protobuf messages.

Currently supported:

- Go

In-progress:

- Python
- Java
- Node

## Usage

```protobuf
syntax = "proto3";
import "github.com/ameliaikeda/protoc-gen-router/proto/router.proto";

servicee myservice {
    option (router).name = "service.foo";
    option (router).short_name = "s-foo";
    
    rpc GETReadUser(GETReadUserRequest) returns (GETReadUSerResponse) {
        option (handler).method = "GET"; // optional with name
        option (handler).path = "/user/read";
    }
}

message GETReadUserRequest {
    string id = 1;
}
message GETReadUserResponse {
    User user = 1;
}

message User {
    string id = 1;
}
```

Now, when you get to code, requesting is pretty easy.

```go
package main

import (
	userproto "mymodule/user/proto"
)

func main() {
	ctx := context.TODO()
    rsp, err := userproto.GETReadUserRequest{
    	Id: "foo",
    }.Response(ctx)
    if err != nil {
    	panic(err)
    }
    
    // rsp.User
}
```