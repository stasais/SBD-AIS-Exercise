package mapred

import (
	"regexp"
	"strings"
	"sync"
)

type MapReduce struct {
}

// Run executes the mapreduce pipeline: map -> shuffle -> reduce
func (mr *MapReduce) Run(input []string) map[string]int {
	// channel for collecting mapper outputs
	mapperResults := make(chan []KeyValue, len(input))
	var wg sync.WaitGroup

	// run map phase concurrently - one goroutine per line
	for _, line := range input {
		wg.Add(1)
		go func(text string) {
			defer wg.Done()
			mapped := mr.wordCountMapper(text)
			mapperResults <- mapped
		}(line)
	}

	// wait for all mappers and close channel
	go func() {
		wg.Wait()
		close(mapperResults)
	}()

	// shuffle phase - group values by key
	intermediate := make(map[string][]int)
	for kvList := range mapperResults {
		for _, kv := range kvList {
			intermediate[kv.Key] = append(intermediate[kv.Key], kv.Value)
		}
	}

	// reduce phase - also run concurrently
	reduceResults := make(chan KeyValue, len(intermediate))
	var reduceWg sync.WaitGroup

	for key, values := range intermediate {
		reduceWg.Add(1)
		go func(k string, vals []int) {
			defer reduceWg.Done()
			reduced := mr.wordCountReducer(k, vals)
			reduceResults <- reduced
		}(key, values)
	}

	go func() {
		reduceWg.Wait()
		close(reduceResults)
	}()

	// collect final results
	results := make(map[string]int)
	for kv := range reduceResults {
		results[kv.Key] = kv.Value
	}

	return results
}

// wordCountMapper splits text into words and emits (word, 1) for each
func (mr *MapReduce) wordCountMapper(text string) []KeyValue {
	// regex to keep only letters (removes special chars and numbers)
	reg := regexp.MustCompile(`[^a-zA-Z]+`)
	cleaned := reg.ReplaceAllString(text, " ")

	// split into words and convert to lowercase
	words := strings.Fields(cleaned)
	result := make([]KeyValue, 0, len(words))

	for _, word := range words {
		word = strings.ToLower(word)
		if word != "" {
			result = append(result, KeyValue{Key: word, Value: 1})
		}
	}

	return result
}

// wordCountReducer sums up all counts for a given word
func (mr *MapReduce) wordCountReducer(key string, values []int) KeyValue {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return KeyValue{Key: key, Value: sum}
}
