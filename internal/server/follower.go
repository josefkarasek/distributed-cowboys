package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/josefkarasek/distributed-cowboys/internal/config"
)

func (f *follower) Start() {
	fmt.Println("Waiting for shootout to start")

	for range f.shootChan {
		err := shootAtTarget(fmt.Sprintf("%s-%d", f.shootoutName, f.c.FindTarget()), f.damage)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if f.c.AmIAlive() {
		// if alive at the end -> last man standing
		fmt.Printf("Winner of this shootout is %s\n", f.name)
	}
}

func (f *follower) start(w http.ResponseWriter, r *http.Request) {
	f.shootChan <- true
}

func (f *follower) stop(w http.ResponseWriter, r *http.Request) {
	close(f.shootChan)
}

func (f *follower) receiveShot(w http.ResponseWriter, r *http.Request) {
	dmg := mux.Vars(r)["damage"]
	i, _ := strconv.ParseInt(dmg, 10, 64)
	remaining := f.c.ReceiveDamage(i)
	fmt.Printf("%s received damage [%d], remaining health [%d]\n", f.name, i, remaining)
}

func (f *follower) stats(w http.ResponseWriter, r *http.Request) {
	new := []config.Cowboy{}
	err := json.NewDecoder(r.Body).Decode(&new)
	if err != nil {
		fmt.Println(err.Error())
	}
	f.c.Update(new)

	data, err := json.Marshal(f.c.Get())
	if err != nil {
		fmt.Println(err.Error())
	}

	url := fmt.Sprintf("http://%s:8080/stats", f.neighbor)
	_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err.Error())
	}
}
