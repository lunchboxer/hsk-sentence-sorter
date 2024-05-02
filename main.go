package main

import (
	"fmt"
	"github.com/lunchboxer/hsk-sentence-sorter/build_dictionary"
	"github.com/lunchboxer/hsk-sentence-sorter/sentence_sort"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		displayHelp()
		return
	}

	command := os.Args[1]
	switch command {
	case "sentences":
		sentencesort.SortSentences()
	case "dictionary":
		builddictionary.BuildDictionary()
		// Call build_dictionary module
		// Implement the functionality for the "dictionary" command
	default:
		displayHelp()
	}
}

func displayHelp() {
	fmt.Println("Usage:")
	fmt.Println("  sentences: Sort sentences based on character levels.")
	fmt.Println("  dictionary: Build a dictionary from character lists.")
	fmt.Println("  [no command]: Display this help menu.")
}
