package main

import (
	"fmt"
	"program/internal/config"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)
}