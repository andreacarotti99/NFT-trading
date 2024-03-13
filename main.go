package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

var (
    mutex                    sync.Mutex
    traitValueCount          = make(map[string]int)
    traitCategoryValueCount  = make(map[string]int)
)


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
        MaxIdleConns:        40, 
        IdleConnTimeout:     90 * time.Second, // Maximum amount of time an idle (keep-alive) connection will remain idle before being closed
        ExpectContinueTimeout: 1 * time.Second, // Prevents the client from waiting too long for this intermediate response.
    },
}

// getToken retrieves a token by its ID (tid) from a specified collection URL (colUrl)
// It returns a pointer to a Token struct populated with the token's attributes
func getToken(tid int, colUrl string) *Token {
	// Constructs the request URL using the base URL, collection URL, and token ID. The url points to a .json containing token info
    url := fmt.Sprintf("%s/%s/%d.json", URL, colUrl, tid)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        logger.Println(string(COLOR_RED), fmt.Sprintf("Error creating request for token %d: ", tid), err, string(COLOR_RESET))
        return &Token{}
    }

	// Sends the HTTP request using the custom HTTP client
    res, err := customClient.Do(req)
    if err != nil {
        logger.Println(string(COLOR_RED), fmt.Sprintf("Error getting token %d: ", tid), err, string(COLOR_RESET))
        return &Token{}
    }
	
	// Ensures the response body is closed after the function returns.
    defer res.Body.Close()

	// Reads the entire response body.
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        logger.Println(string(COLOR_RED), fmt.Sprintf("Error reading response for token %d: ", tid), err, string(COLOR_RESET))
        return &Token{}
    }

	// turns the JSON response into a map to represent the token's attributes.
    attrs := make(map[string]string)
    json.Unmarshal(body, &attrs)

	// Returns a pointer to a Token struct
    return &Token{
        id:    tid,
        attrs: attrs,
    }
}

// this getTokens is never used
// iterate of each element of the collection and call the getToken func
func getTokens(col Collection) []*Token {
	tokens := make([]*Token, col.count)
	for i := 1; i < 10; i++ {
		logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", i), string(COLOR_RESET))
		tokens[i] = getToken(i, col.url)
	}
	return tokens
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

	// Fetch tokens and their metadata for the 'azuki' collection. The function
	if *useWorkerPool {
		// 'GetTokensAndMetadataWithWorkerPool' uses a worker pool pattern to manage concurrency. This is defined in 'workers.go'
		tokens = GetTokensAndMetadataWithWorkerPool(azuki, *maxRoutines) //implementation in workers.go
	} else {
		// 'getTokensAndMetadataWithSemaphore' utilizes a semaphore to limit concurrent
		// HTTP requests, avoiding rate limits or server overload.
		tokens = getTokensAndMetadataWithSemaphore(azuki)
	}	

	// Initialize a priority queue (heap) to store tokens by their rarity scores.
	rarityHeap := NewRarityHeap() 


	// Iterate through all fetched tokens
	for _, token := range tokens {
		if token != nil {

			// Compute the rarity score of the current token
			rarityScore := computeRarity(token)

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

	displayOrderedTopKTokens(rarityHeap)

	// Calculate the total duration
	endTime := time.Now()
	fmt.Printf("Total execution time: %v\n", endTime.Sub(startTime))
}


func printTokenTraitsAndRarity(tokens []*Token) {
    for _, token := range tokens {
        if token != nil {
            rarityScore := computeRarity(token)
            fmt.Printf("Token ID: %d, Rarity: %f, Traits: {\n", token.id, rarityScore)
            for trait, value := range token.attrs {
                fmt.Printf("  \"%s\": \"%s\",\n", trait, value)
            }
            fmt.Println("}")
        }
    }
}


// updateTraitCounts safely updates global maps (traitValueCount and traitCategoryValueCount)
// that track the count of each unique trait value and the count of unique values per trait category
// for a given token. This function is designed to be called concurrently in a multithreaded environment.
func updateTraitCounts(token *Token) {
	// Lock the mutex to ensure exclusive access to the global maps. This prevents concurrent access from multiple goroutines.
    mutex.Lock()

	// Schedule the mutex to be unlocked once the function execution is complete
    defer mutex.Unlock() 

	// Iterate through each attribute (trait and its value) in the token
    for trait, value := range token.attrs {
		
		// Construct a unique key for the trait value by combining the trait name and value.
        traitValueKey := fmt.Sprintf("%s: %s", trait, value)

		// Increment the count for this specific trait value in the traitValueCount map.
        // This tracks how many times each specific trait value appears across all tokens.
        traitValueCount[traitValueKey]++

		// This tracks how many different values each trait has across all tokens.
        if traitValueCount[traitValueKey] == 1 {
            traitCategoryValueCount[trait]++
        }
    }
}

// fetchAndUpdateToken retrieves a token by its ID and collection URL, updates global counts for its traits, and returns the token
func fetchAndUpdateToken(id int, colUrl string) *Token {
	// Retrieve the token from the specified collection URL and ID.
    token := getToken(id, colUrl)
    if token != nil {
		// If the token exists, update the global trait counts based on this token's traits
        updateTraitCounts(token)
    }
    return token
}


// getTokensAndMetadataWithSemaphore concurrently fetches metadata for each token in a given collection
// using a semaphore to limit the number of concurrent fetch operations.
func getTokensAndMetadataWithSemaphore(col Collection) []*Token {

	// Initialize a slice to hold pointers to the Token structs, one for each token in the collection
    tokens := make([]*Token, col.count)

	// Use a WaitGroup to wait for all goroutines launched to finish
    var wg sync.WaitGroup

	// Create a channel to communicate tokens back from goroutines
    // Buffered with col.count to potentially hold all tokens and avoid blocking on send
    tokenChan := make(chan *Token, col.count) 

	// Create a semaphore with a capacity of maxRoutines to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, *maxRoutines)

	// Loop over each token ID in the collection
    for i := 1; i <= col.count; i++ {
        wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore (blocking if full)

        go func(id int) {
            defer wg.Done() // Signal goroutine's completion upon return
			defer func() { <-semaphore }() // Release semaphore after the goroutine finishes
			//logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", id), string(COLOR_RESET))
            
			// Fetch the token's metadata and update the global maps holding the frequency of the attrs.
			token := fetchAndUpdateToken(id, col.url)

			// Send the token through the channel to be collected outside the goroutine.
            tokenChan <- token
        }(i)
    }

	// Launch a goroutine to close the tokenChan once all fetch goroutines have completed.
    go func() {
        wg.Wait()
        close(tokenChan) // Close the channel
    }()

	// Collect tokens from the channel as they arrive and store them in the tokens slice.
    index := 0
    for token := range tokenChan {
        tokens[index] = token
        index++
    }
    return tokens
}

// computes rarity of the given token using traitValueCount and traitCategoryValueCount
func computeRarity(token *Token) float64 {
    var rarity float64
    for trait, value := range token.attrs {
        traitValueKey := fmt.Sprintf("%s: %s", trait, value)
        countWithTraitValue, exists := traitValueCount[traitValueKey]
        if !exists {
            logger.Println("Error: Trait value not found in traitValueCount map")
            continue
        }
        numValuesInCategory, exists := traitCategoryValueCount[trait]
        if !exists {
            logger.Println("Error: Trait category not found in traitCategoryValueCount map")
            continue
        }
        rarity += 1 / (float64(countWithTraitValue) * float64(numValuesInCategory))
    }
    return rarity
}



func displayOrderedTopKTokens(rarityHeap *RarityHeap) {

	length := rarityHeap.Len()

	fmt.Printf("Top %d Rarest Tokens:\n", *topK)

	for i := 0; i < length; i++ {
        element := heap.Pop(rarityHeap).(*RarityScorecard)
        fmt.Printf("Token ID: %d, Rarity: %f\n", element.id, element.rarity)
    }
}

// debug function
func printMaps() {
    // Print traitValueCount map
    fmt.Println("Trait-Value Counts:")
    for traitValue, count := range traitValueCount {
        fmt.Printf("  %s: %d\n", traitValue, count)
    }
    
    // Print traitCategoryValueCount map
    fmt.Println("\nTrait Category Value Counts:")
    for category, valueCount := range traitCategoryValueCount {
        fmt.Printf("  %s: %d\n", category, valueCount)
    }
}