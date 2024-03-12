package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

const URL = "https://go-challenge.skip.money"
const COLLECTION = "azuki"
const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"

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

func printTokenTraits(tokens []*Token) {
	for _, token := range tokens {
		if token != nil {
			fmt.Printf("Token ID: %d, Traits: {\n", token.id)
			for trait, value := range token.attrs {
				fmt.Printf("  \"%s\": \"%s\",\n", trait, value)
			}
		}
	}
}

func main() {
	azuki := Collection{
		count: 10000,
		url:   "azuki1",
	}

	tokens := getTokensConcurrently(azuki)
	printTokenTraits(tokens)
}


func getTokensConcurrently(col Collection) []*Token {
	tokens := make([]*Token, col.count)
	var wg sync.WaitGroup
	tokenChan := make(chan *Token, col.count) // Channel to collect tokens

	for i := 1; i <= col.count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", i), string(COLOR_RESET))
			token := getToken(id, col.url)
			tokenChan <- token // Send the token to the channel
		}(i)
	}

	go func() {
		wg.Wait()
		close(tokenChan) // Close the channel once all goroutines are done
	}()

	// Collect tokens from the channel
	index := 0
	for token := range tokenChan {
		tokens[index] = token
		index++
	}

	return tokens
}