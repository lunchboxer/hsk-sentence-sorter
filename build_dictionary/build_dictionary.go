package builddictionary

import (
	"encoding/json"
	"fmt"
	"github.com/mozillazg/go-pinyin"
	"os"
	"strings"
)

const (
	charactersPinyinFile          = "characters-pinyin.tsv"
	pinyinDictionaryFile          = "pinyin-dictionary.json"
	characterReplacementIndexFile = "character-replacement-index.json"
)

func BuildDictionary() {
	fmt.Println("Compiling pinyin character dictionary...")

	if _, err := os.Stat(charactersPinyinFile); os.IsNotExist(err) {
		fmt.Println("Build Character list by pinyin first")
		makeCharactersPinyinFile()
	}
	// Read the charactersPinyinFile and build the character map
	characterMap, err := buildCharacterMap()
	if err != nil {
		fmt.Printf("Error building character map: %v\n", err)
		return
	}

	fmt.Printf("Found %d characters.\n", len(characterMap))

	// Group characters by pinyin and build the pinyin dictionary
	pinyinGroups := groupCharactersByPinyin(characterMap)

	fmt.Printf("Found %d pinyin groups\n", len(pinyinGroups))

	// Write the pinyin dictionary to pinyinDictionaryFile
	err = writePinyinDictionary(pinyinGroups)
	if err != nil {
		fmt.Printf("Error writing pinyin dictionary: %v\n", err)
		return
	}

	fmt.Printf("Pinyin dictionary built and written to %s.\n", pinyinDictionaryFile)

	// with pinyinGroups and characterMap already in memory, we can merge them into replacement index
	replacementDictionary := buildReplacementDictionary(characterMap, pinyinGroups)

	// Write the replacement dictionary to a JSON file
	err = writeReplacementDictionary(replacementDictionary)
	if err != nil {
		fmt.Printf("Error writing replacement dictionary: %v\n", err)
		return
	}

	fmt.Printf("Replacement dictionary built and written to %s.\n", characterReplacementIndexFile)
}

func buildReplacementDictionary(characterMap map[string]string, pinyinGroups map[string][]string) map[string][]string {
	var noReplacementsCharacters []string
	replacementDictionary := make(map[string][]string)
	for character, pinyin := range characterMap {
		if characters, ok := pinyinGroups[pinyin]; ok {
			// Exclude the character itself from the list of replacements
			var replacements []string
			for _, c := range characters {
				if c != character {
					replacements = append(replacements, c)
				}
			}
			// Check if there are replacements other than the character itself
			if len(replacements) > 0 {
				replacementDictionary[character] = replacements
			} else {
				noReplacementsCharacters = append(noReplacementsCharacters, character)
			}
		}
	}
	if len(noReplacementsCharacters) > 0 {
		fmt.Printf("The following characters have no replacements: %s\n", strings.Join(noReplacementsCharacters, ", "))
	}
	return replacementDictionary
}

func writeReplacementDictionary(replacementDictionary map[string][]string) error {
	data, err := json.Marshal(replacementDictionary)
	if err != nil {
		return err
	}

	return os.WriteFile(characterReplacementIndexFile, data, 0644)
}
func buildCharacterMap() (map[string]string, error) {
	data, err := os.ReadFile(charactersPinyinFile)
	if err != nil {
		return nil, err
	}

	characterMap := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) >= 2 {
			characterMap[fields[0]] = fields[1]
		}
	}

	return characterMap, nil
}

func groupCharactersByPinyin(characterMap map[string]string) map[string][]string {
	pinyinDictionary := make(map[string][]string)
	for character, pinyin := range characterMap {
		pinyinDictionary[pinyin] = append(pinyinDictionary[pinyin], character)
	}
	return pinyinDictionary
}

func writePinyinDictionary(pinyinDictionary map[string][]string) error {
	data, err := json.Marshal(pinyinDictionary)
	if err != nil {
		return err
	}

	return os.WriteFile(pinyinDictionaryFile, data, 0644)
}

func makeCharactersPinyinFile() {

	output := "character\tpinyin\n"

	for i := 1; i <= 7; i++ {
		data, err := os.ReadFile(fmt.Sprintf("data/HSK%d-characters.txt", i))
		if err != nil {
			panic(err)
		}
		characters := strings.Split(string(data), "\n")
		for _, char := range characters {
			// get that character's pinyin and add character and pinyin to the dictionary
			if len(char) == 0 {
				continue
			}
			pinyinStr := pinyin.Pinyin(string(char), pinyin.NewArgs())
			if len(pinyinStr) == 0 {
				fmt.Printf("Unable to find pinyin for %s\n", string(char))
				continue
			} else {
				output += fmt.Sprintf("%s\t%s\n", char, pinyinStr[0][0])
			}
		}
	}

	// Write the output to a TSV file
	err := os.WriteFile(charactersPinyinFile, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Output written to %s\n", charactersPinyinFile)
}
