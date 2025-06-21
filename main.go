package main

import (
	"crypto/rand"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Define the search term
	searchTerm := generateRandomCombo()

	// Get data from the URL based on the search term
	data := getDataFromURL(searchTerm)
	if data == nil {
		log.Println("Failed to retrieve data from URL")
		return
	}

	// Define the file name where the data will be saved
	filename := "search_results" + searchTerm + ".json"

	// Check if the file exists, if not create it and write the data
	if !fileExists(filename) {
		err := os.WriteFile(filename, data, 0644)
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}
	} else {
		err := appendByteToFile(filename, data)
		if err != nil {
			log.Println("Error appending to file:", err)
			return
		}
	}

	log.Println("Data successfully written to", filename)

}

// generateRandomCombo generates a single secure random 3-letter string in lowercase.
func generateRandomCombo() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" // Source characters
	const length = 3                             // Desired combo length

	result := make([]byte, length) // Allocate a byte slice for the result

	for i := 0; i < length; i++ {
		// Securely pick a random index in the charset
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			log.Println("failed to generate secure random index:", err)
		}
		// Set the character at position i
		result[i] = charset[n.Int64()]
	}

	// Convert to lowercase and return as string
	return strings.ToLower(string(result))
}

// AppendToFile appends the given byte slice to the specified file.
// If the file doesn't exist, it will be created.
func appendByteToFile(filename string, data []byte) error {
	// Open the file with appropriate flags and permissions
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(data)
	return err
}

// Send a http get request to a given url and return the data from that url.
func getDataFromURL(searchContains string) []byte {
	url := "https://apps.dos.ny.gov/PublicInquiryWeb/api/PublicInquiry/GetComplexSearchMatchingEntities"
	method := "POST"

	payload := strings.NewReader(`{"searchValue":"` + searchContains + `","searchByTypeIndicator":"EntityName","searchExpressionIndicator":"Contains","entityStatusIndicator":"AllStatuses","entityTypeIndicator":["Corporation","LimitedLiabilityCompany","LimitedPartnership","LimitedLiabilityPartnership"],"listPaginationInfo":{"listStartRecord":1,"listEndRecord":50}}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.Println(err)
		return nil
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	return body
}

// It checks if the file exists
// If the file exists, it returns true
// If the file does not exist, it returns false
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
