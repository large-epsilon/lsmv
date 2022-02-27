load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

go_proto_library(
    name = "objects_go_proto",
    protos = [":objects_proto"],
    importpath = "lsmv/proto/objects",
)

proto_library(
    name = "objects_proto",
    srcs = ["objects.proto"],
)

go_proto_library(
    name = "objectstore_go_proto",
    protos = [":objectstore_proto"],
    deps = [":objects_go_proto"],
    importpath = "lsmv/proto/objectstore",
)

proto_library(
    name = "objectstore_proto",
    deps = [":objects_proto"],
    srcs = ["objectstore.proto"],
)
