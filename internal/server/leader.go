package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/josefkarasek/distributed-cowboys/internal/config"
)

func (l *leader) Start() {
	time.Sleep(5 * time.Second)

	for {
		// start round
		l.startRound()
		// shoot
		err := shootAtTarget(fmt.Sprintf("%s-%d", l.shootoutName, l.c.FindTarget()), l.damage)
		if err != nil {
			fmt.Println(err.Error())
		}

		time.Sleep(1 * time.Second)

		// pass health info further in the ring
		data, err := json.Marshal(l.c.Get())
		if err != nil {
			fmt.Println(err.Error())
		}
		url := fmt.Sprintf("http://%s:8080/stats", l.neighbor)
		_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println(err.Error())
		}

		// wait till ring is synced
		<-l.syncChan

		winner, end := l.isOver()
		if end {
			fmt.Printf("Last man standing is %s\n", winner)
			break
		}
	}

	l.endShootout()

	if l.c.AmIAlive() {
		// if alive at the end -> last man standing
		fmt.Printf("Winner of this shootout is %s\n", l.name)
	}

}

func (l *leader) receiveShot(w http.ResponseWriter, r *http.Request) {
	dmg := mux.Vars(r)["damage"]
	i, _ := strconv.ParseInt(dmg, 10, 64)
	remaining := l.c.ReceiveDamage(i)
	fmt.Printf("%s received damage [%d], remaining health [%d]\n", l.name, i, remaining)
}

func (l *leader) stats(w http.ResponseWriter, r *http.Request) {
	new := []config.Cowboy{}
	err := json.NewDecoder(r.Body).Decode(&new)
	if err != nil {
		fmt.Println(err.Error())
	}
	l.c.Update(new)

	l.syncChan <- true
}

func (l *leader) startRound() error {
	// see who's standing
	var liveCowboys []int
	cowboys := l.c.Get()
	for i := 1; i < l.rounds; i++ {
		if cowboys[i].Health > 0 {
			liveCowboys = append(liveCowboys, i)
		}
	}
	// start round
	for _, cb := range liveCowboys {
		url := fmt.Sprintf("http://%s-%d:8080/start", l.shootoutName, cb)
		_, err := http.Get(url)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func (l *leader) endShootout() error {
	for i := 1; i < l.rounds; i++ {
		url := fmt.Sprintf("http://%s-%d:8080/stop", l.shootoutName, i)
		_, err := http.Get(url)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func (l *leader) isOver() (string, bool) {
	alive := 0
	index := 0
	for i, cb := range l.c.Get() {
		if cb.Health > 0 {
			alive++
			index = i
		}
	}
	if alive == 1 {
		return l.c.Get()[index].Name, true

	}
	return "", false
}
