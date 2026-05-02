package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/onelittlenightmusic/skills-rpg/server"
)

func main() {
	defaultData := filepath.Join(os.Getenv("HOME"), ".mywant-rpg")
	if v := os.Getenv("RPG_DATA_DIR"); v != "" {
		defaultData = v
	}
	defaultStages := "stages"
	if v := os.Getenv("RPG_STAGES_DIR"); v != "" {
		defaultStages = v
	}
	defaultPort := 7100
	if v := os.Getenv("RPG_SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			defaultPort = n
		}
	}

	dataDir := flag.String("data-dir", defaultData, "data directory (current.yaml + saves/)")
	stagesDir := flag.String("stages-dir", defaultStages, "directory containing stage YAMLs")
	port := flag.Int("port", defaultPort, "HTTP listen port")
	reset := flag.Bool("reset", false, "discard current.yaml and bootstrap from stages/")
	flag.Parse()

	cfg := server.Config{DataDir: *dataDir, StagesDir: *stagesDir, Port: *port}
	s, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}
	if *reset {
		if err := s.Reset(); err != nil {
			log.Fatalf("reset: %v", err)
		}
	}

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("rpg-server listening on %s, data=%s stages=%s", addr, *dataDir, *stagesDir)
	if err := http.ListenAndServe(addr, s.Handler()); err != nil {
		log.Fatal(err)
	}
}
