package main

import (
	"log"

	"github.com/danielmiranda22/gator/internal/config"
)

func main() {
	// first read
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to read config:", err)
		return
	}

	// set up and write to disk
	if err := cfg.SetUser("Daniel Miranda"); err != nil {
		log.Fatal("Failed to set user:", err)
		return
	}

	// read again to verify
	cfg, err = config.Read()
	if err != nil {
		log.Fatal("Failed to read updated config:", err)
		return
	}

	log.Println("DB URL:", cfg.DBURL)
	log.Println("Current User Name:", cfg.CurrentUserName)
}
