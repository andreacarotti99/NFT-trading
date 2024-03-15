package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"container/heap"
	"flag"
)


const URL = "https://go-challenge.skip.money"
const COLLECTION = "azuki"
const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"

var (
	maxRoutines = flag.Int("routines", 500, "the maximum number of concurrent goroutines") // Value suggested less than 2500
	topK = flag.Int("topK", 5, "the number of top rarest tokens to display")
	useWorkerPool = flag.Bool("useWorkerPool", false, "use worker pool pattern instead of semaphore for concurrency")
)

var (
    mutex                    sync.Mutex
    traitValueCount          = make(map[string]int)
    traitCategoryValueCount  = make(map[string]int)
)

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

type Token struct {
	id    int
	attrs map[string]string
}

type RarityScorecard struct {
	rarity float64
	id     int
}

type Collection struct {
	count int
	url   string
}


// Define a custom HTTP client with specific configurations
var customClient = &http.Client{
	// Set a timeout for the HTTP client. Determines how long the client should wait
	Timeout: time.Second * 20, 
	Transport: &http.Transport{
		// Keeping idle connections allows for faster subsequent requests to the same host
		// by reusing the existing connection rather than establishing a new one
		DisableKeepAlives: false,
		MaxIdleConns:        40, 
		IdleConnTimeout:     90 * time.Second, // Maximum amount of time an idle (keep-alive) connection will remain idle before being closed
		ExpectContinueTimeout: 1 * time.Second, // Prevents the client from waiting too long for this intermediate response.
    },
}


func main() {
	flag.Parse()

	// Record the start time of the program execution for measurements
	startTime := time.Now()

	// Initialize the azuki collection
	azuki := Collection{
		count: 10000,
		url:   "azuki1",
	}

	var tokens []*Token

	// Fetch tokens and their metadata for the 'azuki' collection. Updating the maps.
	if *useWorkerPool {
		// 'GetTokensAndMetadataWithWorkerPool' uses a worker pool pattern to manage concurrency. This is defined in 'workers.go'
		tokens = GetTokensAndMetadataWithWorkerPool(azuki, *maxRoutines) //implementation in workers.go
	} else {
		// 'getTokensAndMetadataWithSemaphore' utilizes a semaphore to limit concurrent HTTP requests. This is defined in 'semaphores.go'
		tokens = GetTokensAndMetadataWithSemaphore(azuki)
	}	

	// Initialize a priority queue (heap) to store tokens by their rarity scores.
	rarityHeap := NewRarityHeap() 


	// Iterate through all fetched tokens and adding to the heap the rarest tokens
	for _, token := range tokens {
		if token != nil {

			// Compute the rarity score of the current token
			rarityScore := ComputeRarity(token)

			// If the heap does not contain the top K elements yet, add the current token's rarity score.
			// Otherwise, compare the current token's rarity score with the minimum in the heap and replace it if the current token is rarer.
			if rarityHeap.Len() < *topK {
				heap.Push(rarityHeap, &RarityScorecard{rarity: rarityScore, id: token.id})
			} else if rarityScore > (*rarityHeap)[0].rarity {
				heap.Pop(rarityHeap) // Remove the least rare token
				heap.Push(rarityHeap, &RarityScorecard{rarity: rarityScore, id: token.id})
			}
		}
	}

	// function to pop tokens from the heap ordered by rarity
	DisplayOrderedTopKTokens(rarityHeap)

	// Calculate the total execution time
	endTime := time.Now()
	fmt.Printf("Total execution time: %v\n", endTime.Sub(startTime))
}

