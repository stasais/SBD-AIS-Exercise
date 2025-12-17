# Implementation Notes

## Changes from Skeleton

### map_reduce.go
- Added `Run()` method that runs map/shuffle/reduce phases using goroutines
- Added `wordCountMapper()` - uses regex to filter special chars and numbers, splits text into lowercase words
- Added `wordCountReducer()` - sums up counts for each word

### main.go
- Added file reading using `bufio.Scanner`
- Added sorted output of word frequencies

## Test Results

```
=== RUN   Test_Run
--- PASS: Test_Run (0.00s)
=== RUN   Test_Run_Fail
--- PASS: Test_Run_Fail (0.00s)
=== RUN   Test_wordCountMapper
--- PASS: Test_wordCountMapper (0.00s)
=== RUN   Test_wordCountMapper_Fail
--- PASS: Test_wordCountMapper_Fail (0.00s)
=== RUN   Test_wordCountReducer
--- PASS: Test_wordCountReducer (0.00s)
=== RUN   Test_wordCountReducer_Fail
--- PASS: Test_wordCountReducer_Fail (0.00s)
PASS
ok      exc9/mapred     0.330s
```

## Word Count Output

```bash
$ go run main.go | head -51
Word frequencies:
a: 1150
abandoned: 1
abbreviation: 1
abdera: 1
aberration: 1
abide: 5
abides: 1
abideth: 1
abiding: 2
ability: 5
able: 48
ablest: 1
ablutions: 3
abode: 1
abominable: 3
about: 70
above: 13
abroad: 4
abrupted: 1
absence: 3
absent: 3
absolute: 4
absolutely: 9
absorbed: 3
abstain: 1
absurd: 4
abuse: 2
abused: 1
academy: 1
accept: 9
acceptable: 3
acceptation: 1
accepted: 4
accepting: 5
access: 11
accessed: 1
accessible: 1
accessories: 2
accessory: 1
accidentally: 1
accidentary: 1
accidents: 7
acclamation: 1
acclamations: 1
accommodate: 3
accompany: 1
accomplished: 2
accomplishment: 1
accord: 4
accordance: 2
```

Total unique words in Meditations: **6324**
