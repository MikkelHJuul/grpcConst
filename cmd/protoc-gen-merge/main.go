package main

import (
	"github.com/MikkelHJuul/grpcConst/cmd/protoc-gen-merge/merge"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

func main() {
	pgs.Init(
		pgs.DebugEnv("DEBUG"),
	).RegisterModule(
		merge.MakeMerge(),
	).RegisterPostProcessor(
		pgsgo.GoFmt(),
	).Render()
}
