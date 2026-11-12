package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func readDump(t *testing.T) []string {
	t.Helper()
	var known []string
	err := readFile("testdata/dump.json", func(r io.Reader) error {
		var err error
		known, err = ReadNames(r)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return known
}

func TestReadNamesHasCardAndFaceNames(t *testing.T) {
	want := []string{"Alpha Bear", "Bespoke Battlegarb", "Bespoke Bō", "Dáin Ironfoot", "Dóin Ironfoot", "Fire", "Fire // Ice", "Ice"}
	if got := readDump(t); !reflect.DeepEqual(got, want) {
		t.Errorf("names %q, want %q", got, want)
	}
	if _, err := ReadNames(strings.NewReader(`{"name":"x"}`)); err == nil {
		t.Error("a dump that is not an array was read")
	}
}

// TMNT's "Bespoke Bō" is "Bespoke B?" in the datasets.
func TestAMangledNameIsPutBackFromTheDump(t *testing.T) {
	s := NewStats()
	s.card("Bespoke B?").inHand = 600
	s.card("Bespoke B?").inHandWins = 330
	s.card("D?in Ironfoot").inHand = 10 // two dump names fit it
	s.card("Who? Me?").inHand = 10      // none does
	s.card("Alpha Bear").inHand = 5     // a name the dump knows is left alone
	report := s.RepairNames(readDump(t))
	if c, ok := s.cards["Bespoke Bō"]; !ok || c.inHand != 600 || c.inHandWins != 330 {
		t.Errorf("Bespoke Bō: %+v", c)
	}
	if _, ok := s.cards["Bespoke B?"]; ok {
		t.Error("the mangled name is still there")
	}
	for _, n := range []string{"D?in Ironfoot", "Who? Me?", "Alpha Bear"} {
		if _, ok := s.cards[n]; !ok {
			t.Errorf("%q was renamed", n)
		}
	}
	want := []string{
		`"Bespoke B?" is "Bespoke Bō"`,
		`"D?in Ironfoot": 2 names in the dump fit it; left as it is`,
		`"Who? Me?": 0 names in the dump fit it; left as it is`,
	}
	if !reflect.DeepEqual(report, want) {
		t.Errorf("report %q, want %q", report, want)
	}
}

func TestLostLettersMatchesOneNonASCIILetterPerQuestionMark(t *testing.T) {
	for _, c := range []struct {
		dataset, real string
		want          bool
	}{
		{"Bespoke B?", "Bespoke Bō", true},
		{"Bespoke B?", "Bespoke Bo", false},  // an ASCII letter was not lost
		{"Bespoke B?", "Bespoke Bōō", false}, // one letter per "?"
		{"Bespoke B?", "Bespoke Battlegarb", false},
		{"Alpha Bear", "Alpha Bear", true},
	} {
		if got := lostLetters(c.dataset, c.real); got != c.want {
			t.Errorf("lostLetters(%q, %q) = %v", c.dataset, c.real, got)
		}
	}
}
