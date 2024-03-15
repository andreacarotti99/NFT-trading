package main

import (
	"sync"
)

// getTokensAndMetadataWithSemaphore concurrently fetches metadata for each token in a given collection
// using a semaphore to limit the number of concurrent fetch operations.
func GetTokensAndMetadataWithSemaphore(col Collection) []*Token {
	
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
			token := FetchAndUpdateToken(id, col.url)
			
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