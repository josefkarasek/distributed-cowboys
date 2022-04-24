package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/josefkarasek/distributed-cowboys/internal/config"
)

type leader struct {
	name         string
	shootoutName string
	neighbor     string
	damage       int64
	rounds       int
	syncChan     chan bool
	c            *config.Config
}

type follower struct {
	name         string
	shootoutName string
	neighbor     string
	damage       int64
	shootChan    chan bool
	c            *config.Config
}

type Shootout interface {
	Start()
}

func Init(c *config.Config, shootoutName, neighbor string, coordinator bool) Shootout {
	router := mux.NewRouter()
	myself := c.GetMyself()
	var shoot Shootout
	if coordinator {
		l := &leader{
			shootoutName: shootoutName,
			name:         myself.Name,
			damage:       myself.Damage,
			neighbor:     neighbor,
			rounds:       c.Length(),
			syncChan:     make(chan bool, 1),
			c:            c,
		}
		router.Path("/shoot").Queries("damage", "{damage}").HandlerFunc(l.receiveShot)
		router.Path("/stats").HandlerFunc(l.stats).Methods("POST")
		shoot = l
	} else {
		f := &follower{
			shootoutName: shootoutName,
			name:         myself.Name,
			neighbor:     neighbor,
			damage:       myself.Damage,
			c:            c,
			shootChan:    make(chan bool, 1),
		}
		router.Path("/start").HandlerFunc(f.start)
		router.Path("/stop").HandlerFunc(f.stop)
		router.Path("/shoot").Queries("damage", "{damage}").HandlerFunc(f.receiveShot)
		router.Path("/stats").HandlerFunc(f.stats).Methods("POST")
		shoot = f
	}

	go http.ListenAndServe(":8080", router)

	return shoot
}

func shootAtTarget(target string, damage int64) error {
	url := fmt.Sprintf("http://%s:8080/shoot?damage=%d", target, damage)
	_, err := http.Get(url)
	return err
}
