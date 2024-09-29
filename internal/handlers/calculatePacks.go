package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type PackRequest struct {
	Quantity int `json:"quantity"`
}

type PackResponse struct {
	Packs      map[int]int `json:"packs"`
	TotalPacks int         `json:"total_packs"`
}

type PackSizesRequest struct {
	PackSizes []int `json:"pack_sizes"`
}

var (
	PackSizes     = []int{250, 500, 1000, 2000, 5000} // Default pack sizes
	mu            sync.RWMutex
	AuthToken     = "my_secret_token" // Implement something to make this dynamic
	AllowedOrigin = "http://localhost:8080"
	MaxDPQuantity = 100000
)

func CalculatePacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid Content-Type. Expected application/json", http.StatusUnsupportedMediaType)
		return
	}

	var request PackRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Quantity <= 0 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if request.Quantity > math.MaxInt32 {
		http.Error(w, "Quantity exceeds maximum limit", http.StatusBadRequest)
		return
	}

	mu.RLock()
	currentPackSizes := PackSizes
	mu.RUnlock()

	var packDistribution map[int]int
	var totalPacks int

	if request.Quantity <= MaxDPQuantity {
		packDistribution, totalPacks = getOptimalPackDistribution(request.Quantity, currentPackSizes)
	} else {
		packDistribution, totalPacks = greedyPackFallback(request.Quantity, currentPackSizes)
	}

	response := PackResponse{Packs: packDistribution, TotalPacks: totalPacks}
	json.NewEncoder(w).Encode(response)
}

func SetPackSizes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", AllowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid Content-Type. Expected application/json", http.StatusUnsupportedMediaType)
		return
	}

	origin := r.Header.Get("Origin")
	if origin != AllowedOrigin {
		http.Error(w, "Forbidden: Access from this origin is not allowed", http.StatusForbidden)
		return
	}

	token := r.Header.Get("Authorization")
	if !strings.HasPrefix(token, "Bearer ") || strings.TrimPrefix(token, "Bearer ") != AuthToken {
		http.Error(w, "Unauthorized: Invalid authorization token", http.StatusUnauthorized)
		return
	}

	var request PackSizesRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || len(request.PackSizes) == 0 {
		http.Error(w, "Invalid input. Provide a non-empty list of pack sizes.", http.StatusBadRequest)
		return
	}

	for _, size := range request.PackSizes {
		if size <= 0 {
			http.Error(w, "Invalid input. Pack sizes must be positive integers.", http.StatusBadRequest)
			return
		}
	}

	uniqueSizes := uniqueAndSorted(request.PackSizes)

	mu.Lock()
	PackSizes = uniqueSizes
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pack sizes updated successfully"))
}

func uniqueAndSorted(sizes []int) []int {
	sizeMap := make(map[int]bool)
	for _, size := range sizes {
		sizeMap[size] = true
	}

	uniqueSizes := make([]int, 0, len(sizeMap))
	for size := range sizeMap {
		uniqueSizes = append(uniqueSizes, size)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(uniqueSizes)))
	return uniqueSizes
}

func getOptimalPackDistribution(quantity int, packs []int) (map[int]int, int) {
	sort.Ints(packs)

	minPackSize := packs[0]
	maxTotalQuantity := quantity + minPackSize

	dp := make([]int, maxTotalQuantity+1)
	packDistribution := make([]map[int]int, maxTotalQuantity+1)

	for i := range dp {
		dp[i] = math.MaxInt32
		packDistribution[i] = make(map[int]int)
	}

	dp[0] = 0

	for i := 0; i <= maxTotalQuantity; i++ {
		if dp[i] != math.MaxInt32 {
			for _, packSize := range packs {
				newQuantity := i + packSize
				if newQuantity <= maxTotalQuantity {
					if dp[i]+1 < dp[newQuantity] {
						dp[newQuantity] = dp[i] + 1
						packDistribution[newQuantity] = copyMap(packDistribution[i])
						packDistribution[newQuantity][packSize]++
					}
				}
			}
		}
	}

	minOverpack := math.MaxInt32
	minTotalPacks := math.MaxInt32
	minLargestPackSize := math.MaxInt32
	minDistribution := map[int]int{}

	for i := quantity; i <= maxTotalQuantity; i++ {
		if dp[i] != math.MaxInt32 {
			overpack := i - quantity
			totalPacks := dp[i]
			largestPack := maxPackSize(packDistribution[i])

			if overpack < minOverpack {
				minOverpack = overpack
				minTotalPacks = totalPacks
				minLargestPackSize = largestPack
				minDistribution = packDistribution[i]
			} else if overpack == minOverpack {
				if quantity < 500 {
					if largestPack < minLargestPackSize || (largestPack == minLargestPackSize && totalPacks < minTotalPacks) {
						minTotalPacks = totalPacks
						minLargestPackSize = largestPack
						minDistribution = packDistribution[i]
					}
				} else {
					if totalPacks < minTotalPacks || (totalPacks == minTotalPacks && largestPack < minLargestPackSize) {
						minTotalPacks = totalPacks
						minLargestPackSize = largestPack
						minDistribution = packDistribution[i]
					}
				}
			}
		}
	}

	if minTotalPacks == math.MaxInt32 {
		fmt.Printf("No valid solution found for quantity: %d\n", quantity)
		return nil, math.MaxInt32
	}

	return minDistribution, minTotalPacks
}

func maxPackSize(distribution map[int]int) int {
	maxSize := 0
	for size := range distribution {
		if size > maxSize {
			maxSize = size
		}
	}
	return maxSize
}

func greedyPackFallback(quantity int, packs []int) (map[int]int, int) {
	distribution := make(map[int]int)
	remaining := quantity

	sortedPacks := reverseSort(packs)

	for _, packSize := range sortedPacks {
		count := remaining / packSize
		if count > 0 {
			distribution[packSize] = count
			remaining -= packSize * count
		}
	}

	if remaining > 0 && len(sortedPacks) > 0 {
		distribution[sortedPacks[len(sortedPacks)-1]]++
	}

	totalPacks := 0
	for _, count := range distribution {
		totalPacks += count
	}

	return distribution, totalPacks
}

func reverseSort(packs []int) []int {
	sorted := make([]int, len(packs))
	copy(sorted, packs)

	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	return sorted
}

func copyMap(original map[int]int) map[int]int {
	copy := make(map[int]int)
	for key, value := range original {
		copy[key] = value
	}
	return copy
}
