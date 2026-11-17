package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

// The datasets lose some letters outside Latin-1: TMNT's "Bespoke Bō" is
// "Bespoke B?" in every column. The server finds a rated card by name
// (draftratings.Bind), so such a name would rate nothing. RepairNames
// puts the card's real name back from the Scryfall dump.

// ReadNames is every card and face name of a Scryfall dump (a JSON
// array of cards, read as a stream).
func ReadNames(r io.Reader) ([]string, error) {
	dec := json.NewDecoder(r)
	if tok, err := dec.Token(); err != nil {
		return nil, err
	} else if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, errors.New("the dump is not a JSON array")
	}
	seen := map[string]bool{}
	var names []string
	add := func(n string) {
		if n != "" && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	for dec.More() {
		var c struct {
			Name  string `json:"name"`
			Faces []struct {
				Name string `json:"name"`
			} `json:"card_faces"`
		}
		if err := dec.Decode(&c); err != nil {
			return nil, err
		}
		add(c.Name)
		for _, f := range c.Faces {
			add(f.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// lostLetters reports whether a dataset name stands for a real name: the
// same text, but each "?" of the dataset's may be one letter outside
// ASCII in the real one.
func lostLetters(dataset, real string) bool {
	for dataset != "" && real != "" {
		d, dn := utf8.DecodeRuneInString(dataset)
		r, rn := utf8.DecodeRuneInString(real)
		if d != r && (d != '?' || r < utf8.RuneSelf) {
			return false
		}
		dataset, real = dataset[dn:], real[rn:]
	}
	return dataset == "" && real == ""
}

// RepairNames renames each card whose name has a "?" the dump does not
// know to the one dump name it stands for, and reports what it did. A
// name with no match, or more than one, is left as it is and reported.
func (s *Stats) RepairNames(known []string) []string {
	isKnown := make(map[string]bool, len(known))
	for _, n := range known {
		isKnown[n] = true
	}
	var report []string
	var broken []string
	for n := range s.cards {
		if strings.Contains(n, "?") && !isKnown[n] {
			broken = append(broken, n)
		}
	}
	sort.Strings(broken)
	for _, n := range broken {
		var match []string
		for _, k := range known {
			if lostLetters(n, k) {
				match = append(match, k)
			}
		}
		if len(match) != 1 {
			report = append(report, fmt.Sprintf("%q: %d names in the dump fit it; left as it is", n, len(match)))
			continue
		}
		c := s.cards[n]
		delete(s.cards, n)
		into := s.card(match[0])
		into.inHand += c.inHand
		into.inHandWins += c.inHandWins
		into.picks += c.picks
		into.pickSum += c.pickSum
		into.seen += c.seen
		into.seenSum += c.seenSum
		report = append(report, fmt.Sprintf("%q is %q", n, match[0]))
	}
	return report
}
