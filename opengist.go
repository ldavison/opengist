package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"opengist/internal/config"
	"opengist/internal/git"
	"opengist/internal/models"
	"opengist/internal/ssh"
	"opengist/internal/web"
	"os"
	"path/filepath"
)

func initialize() {
	fmt.Println("Opengist v" + config.OpengistVersion)

	configPath := flag.String("config", "", "Path to a config file in YML format")
	flag.Parse()

	if err := config.InitConfig(*configPath); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(config.GetHomeDir()), 0755); err != nil {
		panic(err)
	}

	config.InitLog()

	gitVersion, err := git.GetGitVersion()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if ok, err := config.CheckGitVersion(gitVersion); err != nil {
		log.Fatal().Err(err).Send()
	} else if !ok {
		log.Warn().Msg("Git version may be too old, as Opengist has not been tested prior git version 2.20. " +
			"Current git version: " + gitVersion)
	}

	homePath := config.GetHomeDir()
	log.Info().Msg("Data directory: " + homePath)

	if err := os.MkdirAll(filepath.Join(homePath, "repos"), 0755); err != nil {
		log.Fatal().Err(err).Send()
	}
	if err := os.MkdirAll(filepath.Join(homePath, "tmp", "repos"), 0755); err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msg("Database file: " + filepath.Join(homePath, config.C.DBFilename))
	if err := models.Setup(filepath.Join(homePath, config.C.DBFilename)); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	web.EmbedFS = dirFS
}

func main() {
	initialize()

	go web.Start()
	go ssh.Start()

	select {}
}
