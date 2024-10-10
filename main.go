package main

import (
	"amdzy/gochain/cmd"
	"log"
)

func main() {
	if err := cmd.NewDefaultCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
