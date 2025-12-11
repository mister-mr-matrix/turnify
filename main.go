package main

import (
	"context"
	_ "time/tzdata"

	"github.com/mister-mr-matrix/turnify/src/config"
	"github.com/mister-mr-matrix/turnify/src/router"
)

func main() {
	cfg := config.New()
	rtr := router.New(cfg)

	ctx := context.Background()
	rtr.Start(ctx)
}
