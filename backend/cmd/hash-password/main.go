package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wanglei-123-wl/ika6server/backend/internal/auth"
)

func main() {
	password := ""
	if len(os.Args) > 1 {
		args := os.Args[1:]
		if args[0] == "--" && len(args) > 1 {
			args = args[1:]
		}
		password = args[0]
	} else {
		fmt.Scanln(&password)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash)
}
