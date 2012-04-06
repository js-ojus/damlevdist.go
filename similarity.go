package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const MIN_SET_SIZE = 3
const INIT_INPUT_SIZE = 64

// readData constructs an in-memory vector with all input strings.
func readData(fn string) []string {
	f, err := os.Open(fn)
	if nil != err {
		fmt.Fprintf(os.Stderr, "!! Unable to open the input file: %s; %s\n", os.Args[1], err)
		os.Exit(-1)
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 4*1024)

	v := make([]string, 0, INIT_INPUT_SIZE)

	for {
		by, _, err := r.ReadLine()
		if nil != err {
			break
		}
		s := strings.TrimFunc(string(by), func(c rune) bool {
			if (' ' == c) || ('\t' == c) {
				return true
			}
			return false
		})
		if len(s) > 0 {
			v = append(v, s)
		}
	}

	return v
}

// min answers the smallest element in the given list.
func min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

// distance answers the Damerau-Levenshtein distance between the two
// given strings.  It is not very optimal in that transposed
// characters may be touched again, etc.
//
// In its current form, this function generates garbage every time it
// is invoked.  This can be re-looked at, should GC start slowing the
// process down unacceptably.
func distance(s1, s2 string) int {
	l1 := len(s1)
	l2 := len(s2)

	// First the trivial cases.
	if 0 == l1 {
		return l2
	} else if 0 == l2 {
		return l1
	}

	// Create the dynamic programming matrix.  Note the lengths along
	// both dimensions.  Zeroth row and zeroth column are special.
	tab := make([][]int, l1+2, l1+2)
	for i := 0; i < l1+2; i++ {
		tab[i] = make([]int, l2+2, l2+2)
	}

	// Initialize the matrix with the limit, as needed.
	limit := l1 + l2
	tab[0][0] = limit
	for i := 0; i <= l1; i++ {
		tab[i+1][0] = limit
		tab[i+1][1] = i
	}
	for i := 0; i <= l2; i++ {
		tab[0][i+1] = limit
		tab[1][i+1] = i
	}

	// Main loop.
	for i := 1; i <= l1; i++ {
		cost := -1
		for j := 1; j <= l2; j++ {
			if s1[i-1] == s2[j-1] {
				cost = 0
			} else {
				cost = 1
			}

			tab[i+1][j+1] = min(tab[i][j+1]+1, // Deletion.
				min(tab[i+1][j]+1, // Insertion.
					tab[i][j]+cost)) // Substitution.

			if (i > 1) && (j > 1) && (s1[i-1] == s2[j-2]) && (s1[i-2] == s2[j-1]) {
				tab[i+1][j+1] = min(tab[i+1][j+1],
					tab[i-1][j-1]+cost) // Transposition.
			}
		}
	}

	return tab[l1+1][l2+1]
}

// similarity compares the entire input set, pair-wise, and prints
// their D-L distance, their line numbers and the strings themselves.
func similarity(v []string) {
	l := len(v)
	for i := 0; i < l-1; i++ {
		s1 := v[i]
		for j := i + 1; j < l; j++ {
			s2 := v[j]
			d := distance(s1, s2)
			ml := len(s1) + len(s2)
			fmt.Printf("%.4g\t%d\t%d\t%d\t%s\t%s\n",
				float64(d)/float64(ml), d, i+1, j+1, s1, s2)
		}
	}
}

// updateMinimum checks to see if any of the elements in the vector is
// larger than the given value.  If yes, it places the new element at
// the appropriate index, and removes the highest value from the
// vector.
func updateMinimum(minima []int, d int,
	sv []string, s1 string,
	lnums []int, ln int) ([]int, []string, []int) {
	if 0 == len(minima) {
		minima = append(minima, d)
		sv = append(sv, s1)
		lnums = append(lnums, ln)
		return minima, sv, lnums
	}

	i := -1
	for j, el := range minima {
		if el > d {
			i = j
			break
		}
	}
	if (-1 == i) && (len(minima) < MIN_SET_SIZE) {
		minima = append(minima, d)
		sv = append(sv, s1)
		lnums = append(lnums, ln)
		return minima, sv, lnums
	}

	if i > -1 { // Found a higher element.
		minima = append(minima[:i], append([]int{d}, minima[i:]...)...)
		if len(minima) > MIN_SET_SIZE {
			minima = minima[:len(minima)-1]
		}
		sv = append(sv[:i], append([]string{s1}, sv[i:]...)...)
		if len(sv) > MIN_SET_SIZE {
			sv = sv[:len(sv)-1]
		}
		lnums = append(lnums[:i], append([]int{ln}, lnums[i:]...)...)
		if len(lnums) > MIN_SET_SIZE {
			lnums = lnums[:len(lnums)-1]
		}
	}

	return minima, sv, lnums
}

// refSimilarity compares the test set with the reference set, and
// prints the three most similar reference strings for each test
// string.
func refSimilarity(v1, v2 []string) {
	for i, s2 := range v2 {
		minima := make([]int, 0, MIN_SET_SIZE)
		sv := make([]string, 0, MIN_SET_SIZE)
		lnums := make([]int, 0, MIN_SET_SIZE)

		for j, s1 := range v1 {
			d := distance(s1, s2)
			minima, sv, lnums = updateMinimum(minima, d, sv, s1, lnums, j)
		}

		for k, d := range minima {
			s1 := sv[k]
			ml := len(s1) + len(s2)
			fmt.Printf("%.4g\t%d\t%d\t%d\t%s\t%s\n",
				float64(d)/float64(ml), d, i+1, lnums[k]+1, s2, s1)
		}
		fmt.Printf("----\n")
	}
}

// Print help.
func printHelp() {
	s := `
NAME
    similarity - find and print text similarity between sets of strings

SYNOPSIS
    similarity combfile

    similarity reffile testfile

DESCRIPTION
    similarity is a program that finds the Damerau-Levenshtein distance
    between strings.

    The first form of invocation treats the contents of the file
    'combfile' to be strings that each needs to be compared with all the
    others in the file.

    The second form treats those from the file 'reffile' to be correct
    reference strings, against which those in the file 'testfile' should
    be compared.

    In all cases, the input files should have one string per line.
    Blank lines are ignored.  The program does trim the strings, but
    users should take care of non-printable characters themselves.

    The output will be one line printed for each combination of strings,
    and has the following format:

        pd d tl rl tstr rstr

    where, 'd' is the Damerau-Levenshtein distance between strings 'tstr'
    'rstr'; 'tl' and 'rl' are the line numbers of test string 'tstr' and
    reference string 'rstr', respectively; and 'pd' is calculated as:

        d / (len(tstr) + len(rstr)).

AUTHOR
    JONNALAGADDA Srinivas <js@ojuslabs.com>
`

	fmt.Printf("%s\n", s)
}

// Driver.
func main() {
	switch l := len(os.Args); l {
	default:
		printHelp()

	case 2:
		v := readData(os.Args[1])
		similarity(v)

	case 3:
		v1 := readData(os.Args[1])
		v2 := readData(os.Args[2])
		refSimilarity(v1, v2)
	}
}
