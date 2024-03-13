package main 

import (
	"fmt"
	"container/heap"
)

func DisplayOrderedTopKTokens(rarityHeap *RarityHeap) {

	length := rarityHeap.Len()

	fmt.Printf("Top %d Rarest Tokens:\n", *topK)

	for i := 0; i < length; i++ {
        element := heap.Pop(rarityHeap).(*RarityScorecard)
        fmt.Printf("Token ID: %d, Rarity: %f\n", element.id, element.rarity)
    }
}

// debug function
func PrintMaps() {
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

// debug function
func PrintTokenTraitsAndRarity(tokens []*Token) {
    for _, token := range tokens {
        if token != nil {
            rarityScore := ComputeRarity(token)
            fmt.Printf("Token ID: %d, Rarity: %f, Traits: {\n", token.id, rarityScore)
            for trait, value := range token.attrs {
                fmt.Printf("  \"%s\": \"%s\",\n", trait, value)
            }
            fmt.Println("}")
        }
    }
}