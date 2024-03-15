package main 

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)


// getToken retrieves a token by its ID (tid) from a specified collection URL (colUrl)
// It returns a pointer to a Token struct populated with the token's attributes
func GetToken(tid int, colUrl string) *Token {
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
func GetTokens(col Collection) []*Token {
	tokens := make([]*Token, col.count)
	for i := 1; i < 10; i++ {
		logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", i), string(COLOR_RESET))
		tokens[i] = GetToken(i, col.url)
	}
	return tokens
}


// updateTraitCounts safely updates global maps (traitValueCount and traitCategoryValueCount)
// that track the count of each unique trait value and the count of unique values per trait category
// for a given token. This function is designed to be called concurrently in a multithreaded environment.
func UpdateTraitCounts(token *Token) {
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
func FetchAndUpdateToken(id int, colUrl string) *Token {
	// Retrieve the token from the specified collection URL and ID.
	token := GetToken(id, colUrl)
    if token != nil {
		// If the token exists, update the global trait counts based on this token's traits
		UpdateTraitCounts(token)
    }
    return token
}


