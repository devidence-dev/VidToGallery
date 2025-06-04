package main

import (
	"fmt"
	"log"

	"vidtogallery/pkg/config"
	"vidtogallery/pkg/useragent"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	rotator := useragent.NewRotator(cfg.UserAgent.RandomOrder)

	fmt.Println("Testing User-Agent Rotation:")
	fmt.Printf("Random Order: %v\n\n", cfg.UserAgent.RandomOrder)

	for i := 0; i < 5; i++ {
		ua := rotator.Next()
		fmt.Printf("Request %d: %s\n", i+1, ua)
	}

	fmt.Printf("\nTotal available user agents: %d\n", len(rotator.GetAll()))
}
