package main

import (
	"log"
	"time"
)

func main() {
	start := time.Now()

	//r := new(big.Int)
	//fmt.Println(r.Binomial(1000, 10))
	time.Sleep(2 * time.Second)

	elapsed := time.Since(start)
	log.Printf("Binomial took %s", elapsed)
}
