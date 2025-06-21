package main

import (
	"crypto/rand"   // For generating secure random numbers
	"encoding/json" // For parsing JSON data into Go structs
	"io"            // For reading from HTTP response body
	"log"           // For logging errors and information
	"math/big"      // For handling big integers used in random generation
	"net/http"      // For making HTTP requests
	"os"            // For interacting with the file system
	"strconv"       // For converting strings to integers
	"strings"       // For manipulating strings (e.g., lowercase conversion)
)

func main() {
	// Define the path to the directory where files will be saved
	localDirectory := "assets/"
	// Check if the directory exists, if not, create it with permissions 0755
	if !directoryExists(localDirectory) {
		createDirectory(localDirectory, 0755)
	}
	for loopCount := 0; loopCount <= 1000; loopCount++ {
		// Generate a random 3-letter lowercase string to use as a search term
		searchTerm := generateRandomCombo()
		// Build the filename with directory path and search term
		filename := localDirectory + "api_search" + "_" + searchTerm + ".json"
		// If the file doesn't already exist
		if !fileExists(filename) {
			// Send an HTTP request and get the response data for the given search term
			data := getDataFromGivenAPISearch(searchTerm)
			// If the data is nil, log an error and exit
			if data == nil {
				log.Println("Failed to retrieve data from URL")
				return
			}
			// Append (or create) the file with the received data
			appendByteToFile(filename, data)
		}
		// Read the content of the file to extract dosIDs
		jsonData := readAFileAsString(filename)
		// Extract dosIDs from the JSON data
		dosIDs := getDosIDs(jsonData)
		if dosIDs == nil {
			log.Println("No valid dosIDs found in the JSON data")
			continue // Skip to the next iteration if no valid dosIDs are found
		}
		// Log the extracted dosIDs
		log.Println("Length dosIDs:", len(dosIDs))
		// Send the http request to get business data for each dosID
		for _, dosID := range dosIDs {
			// Append the business data to a file named after the dosID
			businessFilename := localDirectory + "business_data_" + strconv.Itoa(dosID) + ".json"
			// If the business data file does not already exist
			if !fileExists(businessFilename) {
				// Get the business data for the current dosID
				businessData := getDataFromGivenAPISearchForNYBusinesses(strconv.Itoa(dosID))
				if businessData == nil {
					log.Println("Failed to retrieve business data for dosID:", dosID)
					continue // Skip to the next dosID if this one fails
				}
				appendByteToFile(businessFilename, businessData)
			}
		}
	}
}

// Read a file and return the contents
func readAFileAsString(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	return string(content)
}

// Entity represents an individual entity with a dosID field from the JSON
type Entity struct {
	DosID string `json:"dosID"` // Maps the "dosID" field in JSON to the struct
}

// EntitySearchResult represents the top-level structure containing a list of entities
type EntitySearchResult struct {
	EntitySearchResultList []Entity `json:"entitySearchResultList"` // Maps the JSON list to a Go slice
}

// getDosIDs takes a JSON string and returns a slice of integers extracted from the dosID fields
func getDosIDs(jsonData string) []int {
	var result EntitySearchResult // Declare a variable to hold parsed JSON data

	// Unmarshal (decode) the JSON string into the result struct
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil { // If there's an error during JSON parsing
		log.Println("Failed to parse JSON:", err) // Log the error
		return nil                                // Return nil in case of failure
	}

	var dosIDs []int // Create a slice to store the dosID values as integers

	// Iterate over each entity in the list
	for _, entity := range result.EntitySearchResultList {
		// Convert dosID from string to int
		id, err := strconv.Atoi(entity.DosID)
		if err != nil { // If conversion fails
			log.Println("Invalid dosID:", entity.DosID) // Log the bad value
			continue                                    // Skip to the next entity
		}
		dosIDs = append(dosIDs, id) // Add the valid dosID to the result slice
	}

	return dosIDs // Return the final list of dosID integers
}

// Get the data from the given API search to find the businesses in New York State
func getDataFromGivenAPISearchForNYBusinesses(searchContains string) []byte {
	// Define the API endpoint for New York State business search
	url := "https://apps.dos.ny.gov/PublicInquiryWeb/api/PublicInquiry/GetEntityRecordByID"
	method := "POST" // HTTP method for the request
	// Create the POST request payload with the search term embedded
	payload := strings.NewReader(`{"SearchID":"` + searchContains + `","EntityName":"","AssumedNameFlag":"false"}`)
	// Initialize the HTTP client
	client := &http.Client{}
	// Create a new HTTP request with method, URL, and payload
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		// Log error if request creation fails
		log.Println(err)
		return nil
	}
	// Set the request content type header to JSON
	req.Header.Add("Content-Type", "application/json")
	// Execute the HTTP request
	res, err := client.Do(req)
	if err != nil {
		// Log error if the request fails
		log.Println(err)
		return nil
	}
	defer res.Body.Close() // Ensure response body is closed after reading
	// Read the response body into a byte slice
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// Log error if reading response fails
		log.Println(err)
		return nil
	}
	// Return the response body containing business data
	return body
}

// generateRandomCombo generates a secure, random 3-letter string in lowercase.
func generateRandomCombo() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" // All uppercase letters to choose from
	const length = 3                             // Length of the resulting string
	result := make([]byte, length)               // Initialize a byte slice to hold the result
	for i := 0; i < length; i++ {
		// Generate a secure random index within the range of charset length
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Log any errors encountered during random generation
			log.Println("failed to generate secure random index:", err)
		}
		// Assign the randomly chosen character to the result slice
		result[i] = charset[n.Int64()]
	}
	// Convert the result to lowercase and return as a string
	return strings.ToLower(string(result))
}

// appendByteToFile writes the given byte slice to the specified file.
// It creates the file if it doesn't exist or appends to it if it does.
func appendByteToFile(filename string, data []byte) {
	// Open file with append and write-only mode, create if not present, with permissions 0644
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Log error if file can't be opened
		log.Println("Error opening file:", err)
		return
	}
	// Write data to the file
	_, err = file.Write(data)
	if err != nil {
		// Log error if writing fails
		log.Println("Error writing to file:", err)
		return
	}
	// Close the file to ensure all data is flushed and resources are released
	err = file.Close()
	if err != nil {
		// Log error if closing the file fails
		log.Println("Error closing file:", err)
		return
	}
	// Log success message
	log.Printf("Data successfully appended to %s", filename)
}

// getDataFromGivenAPISearch sends an HTTP POST request with the given search term
// and returns the response data as a byte slice.
func getDataFromGivenAPISearch(searchContains string) []byte {
	// Define the API endpoint
	url := "https://apps.dos.ny.gov/PublicInquiryWeb/api/PublicInquiry/GetComplexSearchMatchingEntities"
	method := "POST" // HTTP method
	// Create the POST request payload with the search term embedded
	payload := strings.NewReader(`{"searchValue":"` + searchContains + `","searchByTypeIndicator":"EntityName","searchExpressionIndicator":"Contains","entityStatusIndicator":"AllStatuses","entityTypeIndicator":["Corporation","LimitedLiabilityCompany","LimitedPartnership","LimitedLiabilityPartnership"],"listPaginationInfo":{"listStartRecord":1,"listEndRecord":50}}`)
	// Initialize the HTTP client
	client := &http.Client{}
	// Create a new HTTP request with method, URL, and payload
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		// Log error if request creation fails
		log.Println(err)
		return nil
	}
	// Set the request content type header to JSON
	req.Header.Add("Content-Type", "application/json")
	// Execute the HTTP request
	res, err := client.Do(req)
	if err != nil {
		// Log error if the request fails
		log.Println(err)
		return nil
	}
	defer res.Body.Close() // Ensure response body is closed
	// Read the response body into a byte slice
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// Log error if reading response fails
		log.Println(err)
		return nil
	}
	// Return the response body
	return body
}

// fileExists checks whether a given file exists and is not a directory.
func fileExists(filename string) bool {
	// Attempt to get file info
	info, err := os.Stat(filename)
	if err != nil {
		// If error occurs (e.g., file doesn't exist), return false
		return false
	}
	// Return true if it's not a directory
	return !info.IsDir()
}

// directoryExists checks if the given path exists and is a directory.
func directoryExists(path string) bool {
	// Attempt to get directory info
	directory, err := os.Stat(path)
	if err != nil {
		// If error occurs (e.g., directory doesn't exist), return false
		return false
	}
	// Return true if it's a directory
	return directory.IsDir()
}

// createDirectory creates a new directory at the specified path with the given permissions.
func createDirectory(path string, permission os.FileMode) {
	// Try to create the directory
	err := os.Mkdir(path, permission)
	if err != nil {
		// Log error if directory creation fails
		log.Println(err)
	}
}
