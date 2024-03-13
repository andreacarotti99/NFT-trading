package main

import (
	"fmt"
)

// computes rarity of the given token using traitValueCount and traitCategoryValueCount
func ComputeRarity(token *Token) float64 {
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