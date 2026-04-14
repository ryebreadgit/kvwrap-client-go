{
  pkgs,
  lib,
  config,
  ...
}:
{
  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/packages/
  packages = [
    pkgs.protobuf
    pkgs.protoc-gen-go
    pkgs.protoc-gen-go-grpc
    pkgs.grpcurl
    pkgs.buf
  ];

  # See full reference at https://devenv.sh/reference/options/
}
