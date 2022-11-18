pbtidy
=======

一个小工具，用于修剪 proto 文件编译后的后缀为 `.pb.go` 的文件：

1. 为消息结构体添加`validate`标签，使之可以直接使用 [validator库](https://github.com/go-playground/validator)。
2. 去除枚举定义中的重复部分（该修改是否会导致 protocol buffer 序列化错误待检测）

例如，proto 文件如下：  

```protobuf

enum Done {
  Done_A = 0;
  Done_B = 1;
  Done_C = 2;
}

message Person {
  string name = 1; // 姓名 @{required}
  int64 age = 2; // 年龄 @{required,max=20,min=6} @{这里不会替换}
  bool gender = 3; // 性别
}
```

`protoc` 使用 golang 插件对 proto 文件编译后的原始 `.pb.go` 文件如下：  

```go
type Done int32

const (
	Done_Done_A Done = 0
	Done_Done_B Done = 1
	Done_Done_C Done = 2
)

type Person struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`      // 姓名 @{required}
	Age    int64  `protobuf:"varint,2,opt,name=age,proto3" json:"age,omitempty"`       // 年龄 @{required,max=20,min=6} @{这里不会替换}
	Gender bool   `protobuf:"varint,3,opt,name=gender,proto3" json:"gender,omitempty"` // 性别
}
```

使用本工具对 .pb.go 文件修剪后的效果如下：  

```go 
type Done int32

const (
	Done_A Done = 0
	Done_B Done = 1
	Done_C Done = 2
)

type Person struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty" validate:"required"`      // 姓名 
	Age    int64  `protobuf:"varint,2,opt,name=age,proto3" json:"age,omitempty" validate:"required,max=20,min=6"`       // 年龄  @{这里不会替换}
	Gender bool   `protobuf:"varint,3,opt,name=gender,proto3" json:"gender,omitempty"` // 性别
}
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


