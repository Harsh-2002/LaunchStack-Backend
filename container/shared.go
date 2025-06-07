package container

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GenerateContainerName creates a unique container name for a user instance
func GenerateContainerName(userID uuid.UUID, instanceName string) string {
	// Remove any spaces and special characters from the instance name
	sanitizedName := strings.ToLower(strings.ReplaceAll(instanceName, " ", "-"))
	sanitizedName = strings.ReplaceAll(sanitizedName, "_", "-")
	
	// Create a unique identifier by combining user ID (first 8 chars) and sanitized name
	return fmt.Sprintf("n8n-%s-%s", userID.String()[:8], sanitizedName)
}

// GenerateEasySubdomain generates an easy-to-remember subdomain
func GenerateEasySubdomain(containerName string) string {
	// Lists of cute words for subdomain generation
	animals := []string{"fox", "wolf", "bear", "panda", "koala", "lion", "tiger", "deer", "otter", "seal", 
		"meerkat", "badger", "gecko", "lemur", "sloth", "wombat", "lynx", "fawn"}
	pets := []string{"dog", "cat", "bunny", "hamster", "puppy", "kitten", "rabbit", "ferret", "mouse", "guinea", 
		"beagle", "corgi", "poodle", "husky", "shiba", "pony", "turtle", "budgie", "finch"}
	trees := []string{"oak", "pine", "maple", "birch", "cedar", "aspen", "willow", "spruce", "cherry", "palm", 
		"poplar", "juniper", "walnut", "fir", "elder", "apple", "peach", "plum"}
	fruits := []string{"apple", "pear", "mango", "peach", "berry", "plum", "kiwi", "melon", "grape", "cherry", 
		"lemon", "orange", "lime", "fig", "date", "guava"}
	flowers := []string{"daisy", "tulip", "lily", "rose", "lotus", "violet", "iris", "poppy", "peony", "jasmine", 
		"orchid", "clover", "daffodil", "zinnia"}
	colors := []string{"red", "blue", "green", "gold", "silver", "amber", "rose", "azure", "teal", "coral", 
		"indigo", "ruby", "emerald", "topaz", "jade", "pearl", "mint", "blush"}
	adjectives := []string{"swift", "brave", "happy", "lucky", "sunny", "jolly", "noble", "merry", "calm", "kind", 
		"sweet", "gentle", "witty", "fancy", "clever", "fluffy", "cozy", "snug", "perky", "mellow", "cute"}
	
	// Create two random word lists to pick from, organized to maximize cute combinations
	firstList := []string{}
	firstList = append(firstList, adjectives...)
	firstList = append(firstList, colors...)
	firstList = append(firstList, animals...)
	firstList = append(firstList, pets...)
	
	secondList := []string{}
	secondList = append(secondList, trees...)
	secondList = append(secondList, fruits...)
	secondList = append(secondList, flowers...)
	secondList = append(secondList, animals...)
	secondList = append(secondList, pets...)
	
	// Generate random indices based on hash of containerName
	hash := 0
	for _, c := range containerName {
		hash = (hash*31 + int(c)) % 1000000007
	}
	
	// Pick first and second words using different hash transformations
	firstIndex := (hash % len(firstList))
	secondIndex := ((hash / 100) % len(secondList))
	
	// Ensure we don't have the same word in both positions
	if firstList[firstIndex] == secondList[secondIndex] {
		secondIndex = (secondIndex + 1) % len(secondList)
	}
	
	// Combine words with hyphen
	subdomain := fmt.Sprintf("%s-%s", firstList[firstIndex], secondList[secondIndex])
	
	return subdomain
} 