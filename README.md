# NFT Rarity Calculator

This tool calculates and ranks the rarity of NFTs within a specific collection. It fetches NFT metadata from the defined API endpoint, computes rarity scores and identifies the top rarest NFTs.

## Features

- Fetch NFT metadata from a configurable API endpoint.
- Calculate rarity scores based on the uniqueness of NFT attributes.
- Identify and display the top K rarest NFTs in a collection.

## Requirements

- Go 1.15 or later.


## Design choices
- **Two Maps for traits frequency**:
  - One map for the count of unique values within each trait category.
  - Another for occurrences of each trait-value combination.
- **Parallel Metadata Fetching**:
  - Utilizes goroutines with a configurable level of concurrency.
  - Managed by a semaphore to limit simultaneous HTTP requests, another version with workers is also implemented
- **Rarity Computation is deferred**:
  - Rarity scores are computed iterating over all tokens, after all token metadata is fetched and processed, using the previously fully populated maps.
- **Network**:
  - Custom settings for timeouts, maximum idle connections, and idle connection timeout
- **Heap for Top Rarest Tokens**:
  - Maintains a dynamic collection of the top 5 rarest tokens based on computed rarity scores. After fetching all the data, when we iterate over the tokens and compute the rarity we push the rarest element in the heap. The top K elements are printed out by popping from the heap the elements

## Run the program

#### Run the program
Arguments:
- topK: number of rarest token you want to display (default 5)
- routines: maximum number of concurrent go routines used (suggested value [500-900]) (default 500)
- useWorkerPool: boolean to use the implementation with worker pool (true) or semaphores (default false)
- collection: name of the collection in the defined API endpoint
- numOfTokens: total number of tokens inside the collection

Example:
```
go run . -topK=5 -routines=900 -useWorkerPool=true -collection="azuki1" -numOfTokens=10000
```
The suggested routines were tested with up to 2500 routines, more than that caused network errors.  