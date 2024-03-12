package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"runtime"
	"time"
	"container/heap"
)


const URL = "https://go-challenge.skip.money"
const COLLECTION = "azuki"
const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"

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

func getToken(tid int, colUrl string) *Token {
	url := fmt.Sprintf("%s/%s/%d.json", URL, colUrl, tid)
	res, err := http.Get(url)
	if err != nil {
		logger.Println(string(COLOR_RED), fmt.Sprintf("Error getting token %d :", tid), err, string(COLOR_RESET))
		return &Token{}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Println(string(COLOR_RED), fmt.Sprintf("Error reading response for token %d :", tid), err, string(COLOR_RESET))
		return &Token{}
	}
	attrs := make(map[string]string)
	json.Unmarshal(body, &attrs)
	return &Token{
		id:    tid,
		attrs: attrs,
	}
}


// iterate of each element of the collection and call the getToken func
func getTokens(col Collection) []*Token {
	tokens := make([]*Token, col.count)
	for i := 1; i < 10; i++ {
		// log
		logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", i), string(COLOR_RESET))
		tokens[i] = getToken(i, col.url)
	}
	return tokens
}

// ----

func main() {
	runtime.GOMAXPROCS(2)
	azuki := Collection{
		count: 10000,
		url:   "azuki1",
	}

	tokens := getTokensAndMetadataConcurrently(azuki)

	
	rarityHeap := NewRarityHeap()

	for _, token := range tokens {
		if token != nil {
			rarityScore := computeRarity(token)
			if rarityHeap.Len() < 5 {
				heap.Push(rarityHeap, &RarityScorecard{rarity: rarityScore, id: token.id})
			} else if rarityScore > (*rarityHeap)[0].rarity {
				heap.Pop(rarityHeap)
				heap.Push(rarityHeap, &RarityScorecard{rarity: rarityScore, id: token.id})
			}
		}
	}


	//printTokenTraitsAndRarity(tokens)
	//printMaps()
	fmt.Println("Top 5 Rarest Tokens:")
	for _, scorecard := range *rarityHeap {
		fmt.Printf("Token ID: %d, Rarity: %f\n", scorecard.id, scorecard.rarity)
	}

	
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




// updateTraitCounts updates the global trait count maps in a thread-safe manner
func updateTraitCounts(token *Token) {
    mutex.Lock()
    defer mutex.Unlock()
    for trait, value := range token.attrs {
        traitValueKey := fmt.Sprintf("%s: %s", trait, value)
        traitValueCount[traitValueKey]++
        if traitValueCount[traitValueKey] == 1 {
            traitCategoryValueCount[trait]++
        }
    }
}


// fetchAndUpdateToken combines token fetching and updating counts
func fetchAndUpdateToken(id int, colUrl string) *Token {
    token := getToken(id, colUrl)
    if token != nil {
        updateTraitCounts(token)
    }
    return token
}

// getTokensMetadataConcurrently refactored for better readability
func getTokensAndMetadataConcurrently(col Collection) []*Token {
    tokens := make([]*Token, col.count)
    var wg sync.WaitGroup
    tokenChan := make(chan *Token, col.count) // Channel to collect tokens

	//change to col.count instead of 1000
    for i := 1; i <= col.count; i++ {
		if (i % 1000 == 0){
			time.Sleep(1000 * time.Millisecond)
		}

        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
			logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", id), string(COLOR_RESET))
            token := fetchAndUpdateToken(id, col.url)
            tokenChan <- token
        }(i)
    }

    go func() {
        wg.Wait()
        close(tokenChan)
    }()

    index := 0
    for token := range tokenChan {
        tokens[index] = token
        index++
    }

    return tokens
}

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
