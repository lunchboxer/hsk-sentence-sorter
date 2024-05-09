package sentencesort

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const outputFile = "sentences-by-level.tsv"
const sentencesFile = "data/sentences.txt"
const groupedSentencesFile = "grouped-sentences.json"

func readCharacterList(filePath string, level int) (map[rune]int, error) {
	characterMap := make(map[rune]int)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	for _, char := range string(data) {
		characterMap[char] = level // Set the initial level to 1 for characters in the list
	}

	return characterMap, nil
}
func isRegularChineseCharacter(char rune) bool {
	// Regular Chinese characters typically fall within the unicode range 0x4e00 to 0x9fff
	// Exclude digits and punctuation
	return (char >= 0x4e00 && char <= 0x9fff)
}

func determineSentenceLevel(sentence string, characterMaps ...map[rune]int) int {
	highestLevel := 0

	for _, char := range sentence {
		if !isRegularChineseCharacter(char) {
			continue
		}
		for _, charMap := range characterMaps {
			if level, ok := charMap[char]; ok {
				if level > highestLevel {
					highestLevel = level
				}
			}
		}
	}

	return highestLevel
}

func SortSentences() {
	fmt.Println("Sorting sentences by level of chinese characters...")
	// Read the input sentences
	sentencesData, err := os.ReadFile(sentencesFile)
	if err != nil {
		panic(err)
	}
	sentences := strings.Split(string(sentencesData), "\n")
	fmt.Printf("Found %d sentences\n", len(sentences))

	// Read the character lists
	characterMaps := make([]map[rune]int, 0)
	for i := 1; i <= 7; i++ {
		characterMap, _ := readCharacterList(fmt.Sprintf("data/HSK%d-characters.txt", i), i)
		characterMaps = append(characterMaps, characterMap)
	}

	// Determine and output the levels for each sentence
	output := "sentence\tlevel\n"
	levelMap := make(map[int][]string)
	for _, sentence := range sentences {
		level := determineSentenceLevel(sentence, characterMaps...)
		output += fmt.Sprintf("%s\t%d\n", sentence, level)
		levelMap[level] = append(levelMap[level], sentence)
	}

	// Write the output to a TSV file
	err = os.WriteFile(outputFile, []byte(output), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Output written to %s\n", outputFile)

	// Export the grouped sentences as JSON
	groupedJSON, err := json.MarshalIndent(levelMap, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(groupedSentencesFile, groupedJSON, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Grouped sentences exported to %s\n", groupedSentencesFile)
}
