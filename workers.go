package main 

// this is another implementation using worker pool approach instead of dynamically creating goroutines

func worker(worker_id int, jobs <-chan int, results chan<- *Token, colUrl string) {
	for tokenId := range jobs {
		//logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", j), string(COLOR_RESET))
		token := FetchAndUpdateToken(tokenId, colUrl)
		results <- token
    }
}

func GetTokensAndMetadataWithWorkerPool(col Collection, maxRoutines int) []*Token {
    // Prepare channels for jobs and results
    jobs := make(chan int, col.count)
    results := make(chan *Token, col.count)

    // Start workers
    for w := 1; w <= maxRoutines; w++ {
        go worker(w, jobs, results, col.url)
    }

    // Send jobs
    go func() {
        for i := 1; i <= col.count; i++ {
            jobs <- i
        }
        close(jobs)
    }()

    // Collect results
    tokens := make([]*Token, col.count)
    for i := 1; i <= col.count; i++ {
        token := <-results
        if token != nil {
            tokens[i-1] = token // Assuming token IDs start at 1 and are sequential
        }
    }

    return tokens
}