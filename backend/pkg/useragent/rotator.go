package useragent

import (
	"math/rand"
	"sync"
	"time"
)

type Rotator struct {
	agents    []string
	current   int
	mu        sync.RWMutex
	random    bool
	generator *rand.Rand
}

func NewRotator(random bool) *Rotator {
	return &Rotator{
		agents: []string{
			// Mobile Safari (iOS)
			"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
			"Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
			
			// Chrome Mobile
			"Mozilla/5.0 (Linux; Android 12; SM-G975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Mobile Safari/537.36",
			"Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Mobile Safari/537.36",
			
			// Desktop Chrome
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
			
			// Desktop Safari
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15",
			
			// Firefox
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:102.0) Gecko/20100101 Firefox/102.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0",
		},
		random:    random,
		generator: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Rotator) Next() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.random {
		index := r.generator.Intn(len(r.agents))
		return r.agents[index]
	}

	agent := r.agents[r.current]
	r.current = (r.current + 1) % len(r.agents)
	return agent
}

func (r *Rotator) AddAgent(agent string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = append(r.agents, agent)
}

func (r *Rotator) GetAll() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agents := make([]string, len(r.agents))
	copy(agents, r.agents)
	return agents
}
