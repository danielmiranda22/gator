package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/commands"
	"github.com/danielmiranda22/gator/internal/config"
	"github.com/danielmiranda22/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	programState := &cli.State{
		Cfg: &cfg,
		DB:  database.New(db),
	}

	cmds := cli.Commands{}
	commands.RegisterCommands(&cmds)

	args := os.Args
	if len(args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	cmd := cli.Command{
		Name: args[1],
		Args: args[2:],
	}
	if err := cmds.Run(programState, cmd); err != nil {
		log.Fatal(err)
	}
}
