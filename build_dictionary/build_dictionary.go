package builddictionary

import (
	"encoding/json"
	"fmt"
	"github.com/mozillazg/go-pinyin"
	"os"
	"strings"
)

const (
	charactersPinyinFile                  = "characters-pinyin.tsv"
	charactersPinyinFileAlt               = "characters-pinyin.json"
	pinyinDictionaryFile                  = "pinyin-dictionary.json"
	characterReplacementIndexFile         = "character-replacement-index.json"
	charactersGroupedByPinyinAndLevelFile = "characters-grouped-by-pinyin-and-level.json"
	characterWordMapFile                  = "character-word.json"
)

func BuildDictionary() {
	fmt.Println("Compiling pinyin character dictionary...")

	makeCharactersPinyinFile()
	// Read the charactersPinyinFile and build the character map
	characterMap, err := buildCharacterMap()
	if err != nil {
		fmt.Printf("Error building character map: %v\n", err)
		return
	}

	fmt.Printf("Found %d characters.\n", len(characterMap))

	writeCharactersPinyinFileAlt(characterMap)

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

	characterWordMap, err := getCharacterWordMap()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(len(characterWordMap))

	err = writeCharacterWordMap(characterWordMap)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Character word map built and written to %s.\n", characterWordMapFile)
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

// write characterMap to JSON
func writeCharactersPinyinFileAlt(characterMap map[string]string) error {
	data, err := json.Marshal(characterMap)
	if err != nil {
		return err
	}
	return os.WriteFile(charactersPinyinFileAlt, data, 0644)
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
	pinyinLevelMap := make(map[string]map[int][]string)

	for level := 1; level <= 7; level++ {
		data, err := os.ReadFile(fmt.Sprintf("data/HSK%d-characters.txt", level))
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
				// populate the pinyinLevelMap with the character, pinyin, and level
				pinyinString := pinyinStr[0][0]
				if _, ok := pinyinLevelMap[pinyinString]; !ok {
					pinyinLevelMap[pinyinString] = make(map[int][]string)
				}
				pinyinLevelMap[pinyinString][level] = append(pinyinLevelMap[pinyinString][level], char)
				output += fmt.Sprintf("%s\t%s\n", char, pinyinString)
			}
		}
	}

	// Write the output to a TSV file
	err := os.WriteFile(charactersPinyinFile, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Output written to %s\n", charactersPinyinFile)

	data, err := json.Marshal(pinyinLevelMap)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(charactersGroupedByPinyinAndLevelFile, data, 0644)
	if err != nil {
		panic(err)
	}
}

// returns a list of strings
func listAllCharacters() []string {
	allCharacters := make([]string, 0)
	for level := 1; level <= 7; level++ {
		data, err := os.ReadFile(fmt.Sprintf("data/HSK%d-characters.txt", level))
		if err != nil {
			panic(err)
		}
		characters := strings.Split(string(data), "\n")
		allCharacters = append(allCharacters, characters...)
	}
	return allCharacters
}

// this function loops through `HSK %d - meanings.tsv` with integer for each of 7 levels.
// It returns a map of words in the 2nd column to their meanings in the 4th column
// ignores the other columns
// group words by level
// this function loops through `HSK %d - meanings.tsv` with integer for each of 7 levels.
// It returns a map of level to a map of words to their meanings
func getWordMeaningMap() (map[int]map[string]string, error) {
	wordMeaningMap := make(map[int]map[string]string)

	for level := 1; level <= 7; level++ {
		data, err := os.ReadFile(fmt.Sprintf("data/HSK%d-meanings.tsv", level))
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(data), "\n")

		levelMap := make(map[string]string)
		for _, line := range lines[1:] { // skip header line
			fields := strings.Split(line, "\t")
			if len(fields) >= 4 {
				levelMap[fields[1]] = fields[3]
			}
		}
		wordMeaningMap[level] = levelMap
	}
	return wordMeaningMap, nil
}

func getCharacterWordMap() (map[string]map[int]map[string]string, error) {
	wordMeaningMap, err := getWordMeaningMap()
	if err != nil {
		return nil, err
	}

	allCharacters := listAllCharacters()

	characterWordMap := make(map[string]map[int]map[string]string)

	// for each character, find all words that contain that character
	for _, char := range allCharacters {
		if len(char) == 0 {
			continue
		}
		if _, ok := characterWordMap[char]; !ok {
			characterWordMap[char] = make(map[int]map[string]string)
		}

		for level, levelMap := range wordMeaningMap {
			wordMapForLevel := make(map[string]string)
			for word, meaning := range levelMap {
				if strings.Contains(word, char) {
					wordMapForLevel[word] = meaning
				}
			}
			if len(wordMapForLevel) > 0 {
				characterWordMap[char][level] = wordMapForLevel
			}
		}
	}

	return characterWordMap, nil
}

func writeCharacterWordMap(characterWordMap map[string]map[int]map[string]string) error {
	data, err := json.Marshal(characterWordMap)
	if err != nil {
		return err
	}

	err = os.WriteFile(characterWordMapFile, data, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Output written to %s\n", characterWordMapFile)
	return nil
}
