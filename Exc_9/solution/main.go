package main

import (
	"bufio"
	"exc9/mapred"
	"fmt"
	"log"
	"os"
	"sort"
)

func main() {
	// read meditations.txt into a slice of strings
	file, err := os.Open("res/meditations.txt")
	if err != nil {
		log.Fatal("could not open file:", err)
	}
	defer file.Close()

	var text []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			text = append(text, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("error reading file:", err)
	}

	// run mapreduce
	var mr mapred.MapReduce
	results := mr.Run(text)

	// sort keys for nice output
	keys := make([]string, 0, len(results))
	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// print word frequencies
	fmt.Println("Word frequencies:")
	for _, k := range keys {
		fmt.Printf("%s: %d\n", k, results[k])
	}
	fmt.Printf("\nTotal unique words: %d\n", len(results))
}
