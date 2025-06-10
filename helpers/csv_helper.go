package helpers

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"regexp"
	"strings"
)

func Write(fileName string, rows [][]string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer file.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	w := csv.NewWriter(file)
	defer w.Flush()

	w.Comma = ';'
	err = w.WriteAll(rows)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func ToSnakeCase(text string) string {
	words := strings.Split(text, " ")

	var result []string
	for _, word := range words {
		result = append(result, strings.ToLower(strings.TrimSpace(word)))
	}

	return strings.Join(result, "_")
}

func ToCapitalize(text string) string {
	words := strings.Split(text, " ")

	var result []string
	for _, word := range words {
		result = append(result, cases.Title(language.English).String(cases.Lower(language.English).String(word)))
	}

	return strings.Join(result, "_")
}

func ReadFromCSV(fileName string) ([][]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	// ReadAll reads all the records from the CSV file
	// and Returns them as slice of slices of string
	// and an error if any
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		return nil, err
	}

	return records, nil

}

// Metni normalize edip, Türkçe karakterleri çevirip ve birleştiren fonksiyon
func Normalize(input string) string {
	replacements := map[string]string{
		"ç": "c", "Ç": "C",
		"ğ": "g", "Ğ": "G",
		"ı": "i", "I": "I",
		"İ": "I", "ö": "o",
		"Ö": "O", "ş": "s",
		"Ş": "S", "ü": "u",
		"Ü": "U",
	}

	input = strings.ToLower(input)

	for tr, en := range replacements {
		input = strings.ReplaceAll(input, tr, en)
	}

	reg, _ := regexp.Compile(`[^a-z0-9\s]+`)
	input = reg.ReplaceAllString(input, "")

	input = strings.TrimSpace(input)

	return input
}
