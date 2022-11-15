pbtidy
=======

一个小工具，用于修剪 proto 文件编译后的后缀为 `.pb.go` 的文件，使之可以直接使用 [validator库](https://github.com/go-playground/validator)。

例如，proto 文件如下：  

```protobuf
// ...

message Person {
  string name = 1; // 姓名 @{required}
  int64 age = 2; // 年龄 @{required,max=20,min=6} @{这里不会替换}
  bool gender = 3; // 性别
}

// ...
```

`protoc` 使用 golang 插件对 proto 文件编译后的 `.pb.go` 文件如下：  

```go
// ...

type Person struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`      // 姓名 @{required}
	Age    int64  `protobuf:"varint,2,opt,name=age,proto3" json:"age,omitempty"`       // 年龄 @{required,max=20,min=6} @{这里不会替换}
	Gender bool   `protobuf:"varint,3,opt,name=gender,proto3" json:"gender,omitempty"` // 性别
}

// ...
```

使用本工具对 .pb.go 文件修剪后的效果如下：  

```go 
// ...

type Person struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty" validate:"required"`      // 姓名 
	Age    int64  `protobuf:"varint,2,opt,name=age,proto3" json:"age,omitempty" validate:"required,max=20,min=6"`       // 年龄  @{这里不会替换}
	Gender bool   `protobuf:"varint,3,opt,name=gender,proto3" json:"gender,omitempty"` // 性别
}

// ...
```

Usage && Doc
------------

example:  

```shell
$ pbtidy -dir="/data/testdata" -rec=false
```

doc:  
```shell
$ pbtidy --help

>  -dir string                              
>        指定的目录                         
>  -rec                                     
>        是否递归处理子文件夹 (default true)
```


