package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

func checkUniqueLetters(w string) bool {
	var letters [5]rune
	for i, c := range w {
		for _, l := range letters {
			if l == c {
				return false
			}
			letters[i] = c
		}
	}
	return true
}

func filterDuplicates(wns []uint32) []uint32 {
	var unique_values []uint32
	for _, n := range wns {
		is_duplicate := false
		for _, v := range unique_values {
			if n == v {
				is_duplicate = true
				break
			}
		}
		if !is_duplicate {
			unique_values = append(unique_values, n)
		}
	}
	return unique_values
}

func encodeWord(w string) uint32 {
	var char_values [5]uint32
	for i, c := range w {
		char_values[i] = 1 << uint32(rune(c)-97)
	}
	var word_value uint32
	for _, c := range char_values {
		word_value += c
	}
	return word_value
}

func encodeAllWords(words []string) []uint32 {
	encoded_words := make([]uint32, len(words))
	for i, w := range words {
		encoded_words[i] = encodeWord(w)
	}
	return encoded_words
}

func mapEncodedWords(original_words []string, encoded_words []uint32) map[uint32][]string {
	var word_num_map = make(map[uint32][]string)
	for i, w := range encoded_words {
		word_num_map[w] = append(word_num_map[w], original_words[i])
	}
	return word_num_map
}

func filterWords(words []uint32, char uint32) []uint32 {
	var filtered_words []uint32
	for _, w := range words {
		if w&char == 0 {
			filtered_words = append(filtered_words, w)
		}
	}
	return filtered_words
}

func makeAlphabet() [26]uint32 {
	var alphabet [26]uint32
	for i := 0; i < 26; i++ {
		alphabet[i] = 1 << i
	}
	return alphabet
}

func reverseAlphabet(alphabet [26]uint32, words []uint32) [26]uint32 {

	freqs := make([]int, 26)
	for i, a := range alphabet {
		for _, w := range words {
			if a&w == a {
				freqs[i] += 1
			}
		}
	}
	freqs_sorted := make([]int, 26)
	copy(freqs_sorted, freqs)
	sort.Ints(freqs_sorted)

	var reverse [26]uint32
	for i, fs := range freqs_sorted {
		for j, f := range freqs {
			skip := false
			if f == fs {
				a := alphabet[j]
				for _, c := range reverse[:i] {
					if a == c {
						skip = true
					}
				}
				if skip == false {
					reverse[i] = a
				}
			}
		}
	}
	return reverse
}

// filterAlphabet removes one letter from the alphabet
func filterAlphabet(alphabet [26]uint32, i int) [25]uint32 {
	var filtered [25]uint32
	count := 0
	for j, a := range alphabet {
		if j != i {
			filtered[count] = a
			count += 1
		}
	}
	return filtered
}

// findCharsInWord returns the all the characters for the word passed in the order of the alphabet passed
func findCharsInWord(w uint32, alphabet [26]uint32) []uint32 {
	var chars []uint32
	for _, a := range alphabet {
		if a&w != 0 {
			chars = append(chars, a)
		}
	}
	return chars
}

// findMinChars finds the char with the lowest overall frequency for each word passed
func findMinChars(words []uint32, alphabet [26]uint32) []uint32 {
	min_chars := make([]uint32, len(words))
	for i, w := range words {
		min_chars[i] = findCharsInWord(w, alphabet)[0]
	}
	return min_chars
}

func iteratePossibleSolutions(solution []uint32, word_num_map map[uint32][]string) [][]string {
	var combinations [][]string
	for _, a := range word_num_map[solution[0]] {
		for _, b := range word_num_map[solution[1]] {
			for _, c := range word_num_map[solution[2]] {
				for _, d := range word_num_map[solution[3]] {
					for _, e := range word_num_map[solution[4]] {
						combinations = append(combinations, []string{a, b, c, d, e})
					}
				}
			}
		}
	}
	return combinations
}

func writeOutput(file_name string, results [][]uint32, word_num_map map[uint32][]string) {

	f, _ := os.Create(file_name)
	out := bufio.NewWriter(f)
	for _, r := range results {
		to_write := iteratePossibleSolutions(r, word_num_map)
		for _, row := range to_write {
			for _, word := range row {
				out.WriteString(word)
				out.WriteByte(',')
			}
			out.WriteByte('\n')
		}
	}
	out.Flush()
	f.Close()
}

func readFile(f string) []string {
	file, err := os.Open("words_alpha.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		w := scanner.Text()
		if len(w) == 5 {
			if checkUniqueLetters(w) {
				words = append(words, w)
			}
		}
	}
	return words
}

// findNextWord finds the next set words that creates a unique character set for each partial solution passed
func findNextWord(results [][]uint32, alphabet_filtered [25]uint32, words_filtered []uint32, min_chars []uint32) [][]uint32 {
	// combine the words to help find the next missing letter in order of overall frequency
	next_min_char := make([]uint32, len(results))
	combined_words := make([]uint32, len(results))
	for i, r := range results {
		var combined_word uint32
		for _, w := range r {
			combined_word += w
		}
		combined_words[i] = combined_word
		for _, a := range alphabet_filtered {
			if a&combined_word == 0 {
				next_min_char[i] = a
				// stop after the first letter is found
				break
			}
		}
	}
	next_min_char_unique := filterDuplicates(next_min_char)

	// pre-filter the next potential words to explore based on words that contain a minimum char in the set of next min chars
	var next_words []uint32
	for i, mc := range min_chars {
		for _, nmc := range next_min_char_unique {
			if mc == nmc {
				next_words = append(next_words, words_filtered[i])
				break
			}
		}
	}

	var solutions [][]uint32
	for j, c := range combined_words {
		for _, n := range next_words {
			// check that the new word and combined word have no letters in common
			// check that the new word contains the relavent next min char
			if (n&c == 0) && (next_min_char[j]&n != 0) {
				solutions = append(solutions, append([]uint32{n}, results[j]...))
			}
		}
	}
	return solutions
}

func main() {
	start := time.Now()

	raw_words := readFile("words_alpha.txt")
	words := encodeAllWords(raw_words)
	word_num_map := mapEncodedWords(raw_words, words)
	unique_words := filterDuplicates(words)
	alphabet := makeAlphabet()
	alphabet = reverseAlphabet(alphabet, words)

	mu := sync.Mutex{}
	var wg sync.WaitGroup
	var all_results [][]uint32
	// remove one char at a time and find solutions that fully utilise the remaining chars
	for i := 0; i < 26; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// filter alphabet and word list to exclude values that contain skipped character
			alphabet_filtered := filterAlphabet(alphabet, i)
			words_filtered := filterWords(unique_words, alphabet[i])
			min_chars := findMinChars(words_filtered, alphabet)

			// the first potential solutions must have a min char of the lowest overall frequency
			var results [][]uint32
			for j, c := range min_chars {
				if c == alphabet_filtered[0] {
					results = append(results, []uint32{words_filtered[j]})
				}
			}

			// solve words 2-5
			for j := 0; j < 4; j++ {
				if len(results) > 0 {
					results = findNextWord(results, alphabet_filtered, words_filtered, min_chars)
				} else {
					break
				}
			}

			if len(results) > 0 {
				mu.Lock()
				all_results = append(all_results, results...)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	writeOutput("solutions.csv", all_results, word_num_map)
	fmt.Println("Total time (seconds):", time.Now().Sub(start).Seconds())
}
