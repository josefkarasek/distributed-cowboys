package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	stdlog "log"
	"os"

	"github.com/josefkarasek/distributed-cowboys/internal/config"
	"github.com/josefkarasek/distributed-cowboys/internal/server"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

const (
	configMountPath = "/var/run/cowboys/config"
)

var (
	log logr.Logger
)

func main() {
	var name, shootoutName, neighbor string
	var coordinator bool
	var verbosity int
	flag.StringVar(&name, "name", "", "The name of the node.")
	flag.StringVar(&shootoutName, "shootout-name", "", "The name of this shootout.")
	flag.StringVar(&neighbor, "neighbor", "", "The neighbor of this node.")
	flag.BoolVar(&coordinator, "coordinator", false, "If this node's role is coordinator")
	flag.IntVar(&verbosity, "v", 0, "Log verbosity")
	flag.Parse()

	stdr.SetVerbosity(verbosity)
	log = stdr.NewWithOptions(stdlog.New(os.Stderr, "", stdlog.LstdFlags), stdr.Options{LogCaller: stdr.All})
	log = log.WithName("distributed-cowboys")

	log.Info("Starting up", "Name", name, "Neighbor", neighbor, "Coordinator", coordinator)

	bytes, err := ioutil.ReadFile(configMountPath)
	if err != nil {
		panic(err)
	}
	var cowboys []config.Cowboy
	if err := json.Unmarshal(bytes, &cowboys); err != nil {
		panic(err)
	}

	config := &config.Config{}
	config.Init(name, cowboys)

	for _, c := range cowboys {
		log.Info("Config loaded", "Name", c.Name, "Health", c.Health, "Damage", c.Damage)
	}

	shootout := server.Init(config, shootoutName, neighbor, coordinator)
	shootout.Start()

	os.Exit(0)
}
