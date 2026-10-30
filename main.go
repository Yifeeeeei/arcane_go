package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "gen" {
		b, err := os.ReadFile("template_mass_producer_params_xlsx.json")
		if err != nil {
			panic(err)
		}
		var p MassProducerParams
		if err = json.Unmarshal(b, &p); err != nil {
			panic(err)
		}
		b, _ = json.MarshalIndent(p, "", "    ")
		if err = os.WriteFile("mass_producer_params_xlsx.json", append(b, '\n'), 0644); err != nil {
			panic(err)
		}
		return
	}
	paramPath := "mass_producer_params_xlsx.json"
	if len(args) > 0 && args[0] == "new" {
		paramPath = "new_mass_producer_params.json"
	}
	if len(args) > 0 && args[0] == "render" {
		paramPath = "new_mass_producer_params.json"
	}
	if len(args) > 1 && args[0] == "validate" {
		paramPath = args[1]
	}
	m, err := NewMassProducer(paramPath)
	if err != nil {
		panic(err)
	}
	start := time.Now()
	if len(args) > 0 && args[0] == "validate" {
		if err = m.Validate(); err != nil {
			panic(err)
		}
		fmt.Printf("validated %d cards\n", len(m.AllCardInfos))
		return
	}
	if len(args) > 0 && args[0] == "render" {
		if len(args) < 2 {
			panic("usage: go run . render CARD_ID [CARD_ID ...]")
		}
		if err = m.ProduceIDs(args[1:]); err != nil {
			panic(err)
		}
		fmt.Printf("rendered %d cards\n", len(args)-1)
		return
	}
	if len(args) > 0 && args[0] == "render-old" {
		if len(args) < 2 {
			panic("usage: go run . render-old CARD_ID [CARD_ID ...]")
		}
		if err = m.ProduceIDs(args[1:]); err != nil {
			panic(err)
		}
		fmt.Printf("rendered %d legacy cards\n", len(args)-1)
		return
	}
	if err = m.Produce(); err != nil {
		panic(err)
	}
	fmt.Println("Time elapsed:", time.Since(start))
}
