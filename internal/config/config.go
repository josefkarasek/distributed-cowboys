package config

import (
	"math/rand"
	"sync"
	"time"
)

type Config struct {
	sync.Mutex
	// cowboys map[string]Cowboy
	cowboys []Cowboy
	myName  string
}

type Cowboy struct {
	Name   string `json:"name,omitempty"`
	Health int64  `json:"health,omitempty"`
	Damage int64  `json:"damage,omitempty"`
}

func (c *Config) Init(name string, cw []Cowboy) {
	c.Lock()
	defer c.Unlock()
	c.myName = name
	c.cowboys = cw
}

func (c *Config) GetMyself() Cowboy {
	c.Lock()
	defer c.Unlock()
	for _, cb := range c.cowboys {
		if cb.Name == c.myName {
			return cb
		}
	}
	return Cowboy{}
}

func (c *Config) SetMyself(new Cowboy) {
	c.Lock()
	defer c.Unlock()
	for i, cb := range c.cowboys {
		if cb.Name == c.myName {
			c.cowboys[i] = new
			return
		}
	}
}

func (c *Config) Get() []Cowboy {
	c.Lock()
	defer c.Unlock()
	return c.cowboys
}

func (c *Config) Set(new []Cowboy) {
	c.Lock()
	defer c.Unlock()
	c.cowboys = new
}

// --- Helper functions ---

func (c *Config) Length() int {
	cowboys := c.Get()
	return len(cowboys)
}

func (c *Config) ReceiveDamage(damage int64) int64 {
	me := c.GetMyself()
	newHealth := me.Health - damage
	if newHealth < 0 {
		newHealth = 0
	}
	me.Health = newHealth
	c.SetMyself(me)
	return me.Health
}

func (c *Config) Update(new []Cowboy) {
	me := c.GetMyself()
	for i, cb := range new {
		if cb.Name == me.Name {
			new[i] = me
		}
	}
	c.Set(new)
}

func (c *Config) FindTarget() int {
	cowboys := c.Get()
	var targets []int
	for i, cb := range cowboys {
		if cb.Name != c.myName && cb.Health > 0 {
			targets = append(targets, i)
		}
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	index := r1.Intn(100) % len(targets)
	return targets[index]
}

func (c *Config) AmIAlive() bool {
	me := c.GetMyself()
	return me.Health > 0
}
