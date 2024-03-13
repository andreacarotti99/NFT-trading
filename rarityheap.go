package main 

import (
	"container/heap"
)

type RarityHeap []*RarityScorecard

func (h RarityHeap) Len() int { 
	return len(h) 
}

func (h RarityHeap) Less(i, j int) bool {
	 return h[i].rarity < h[j].rarity 
}

func (h RarityHeap) Swap(i, j int) {
	 h[i], h[j] = h[j], h[i] 
}

func (h *RarityHeap) Push(x interface{}) {
	*h = append(*h, x.(*RarityScorecard))
}

func (h *RarityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewRarityHeap() *RarityHeap {
	var rh RarityHeap
	heap.Init(&rh)
	return &rh
}


