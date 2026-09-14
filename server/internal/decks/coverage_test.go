package decks

import (
	"fmt"
	"strings"
	"testing"
)

// TestCoverageAccountsForEveryCard — the arithmetic has to close, or
// the picker's headline sentence ("99 cards, all fully implemented")
// is describing a different deck than the one being installed.
func TestCoverageAccountsForEveryCard(t *testing.T) {
	covs := Coverages(nil)
	if len(covs) != len(All()) {
		t.Fatalf("Coverages returned %d profiles for %d decks", len(covs), len(All()))
	}
	for _, d := range All() {
		cov, ok := covs[d.ID]
		if !ok {
			t.Fatalf("%s: no coverage profile", d.ID)
		}
		if got, want := cov.Full+cov.Caveats+cov.Unreviewed, cov.Cards; got != want {
			t.Errorf("%s: %d graded cards but Cards = %d", d.ID, got, want)
		}
		var basics, nonbasics int
		for _, c := range d.Cards() {
			if c.Basic {
				basics++
			} else {
				nonbasics++
			}
		}
		if cov.Cards != nonbasics {
			t.Errorf("%s: Cards = %d, deck has %d non-basic rows", d.ID, cov.Cards, nonbasics)
		}
		if cov.Basics != basics {
			t.Errorf("%s: Basics = %d, deck has %d basic rows", d.ID, cov.Basics, basics)
		}
		if got, want := len(cov.Imperfect), cov.Caveats+cov.Unreviewed+cov.Unregistered; got != want {
			t.Errorf("%s: %d imperfect cards listed, %d counted", d.ID, got, want)
		}
		if cov.Complete() != (len(cov.Imperfect) == 0) {
			t.Errorf("%s: Complete() = %v with %d imperfect cards", d.ID, cov.Complete(), len(cov.Imperfect))
		}
	}
}

// TestCoverageReportsNoUnregisteredCards restates the build-failing
// coverage test as the promise the PICKER makes. decks_test.go already
// fails if a card stops resolving to a Spec; this fails if that ever
// stops being true of the numbers the player is shown, which is the
// half a player can actually see.
func TestCoverageReportsNoUnregisteredCards(t *testing.T) {
	for id, cov := range Coverages(nil) {
		if cov.Unregistered != 0 {
			t.Errorf("%s: %d card(s) with no registered effects.Spec; the picker would be advertising a deck the engine cannot play", id, cov.Unregistered)
		}
	}
}

// TestEveryCaveatCardSaysWhy — a caveat count with no readable reason
// is worse than no count: it tells a player something is wrong and
// gives them nothing to judge it by. An `unreviewed` card is the one
// case with nothing to say, and it is labelled as such rather than
// shown as an empty caveat.
func TestEveryCaveatCardSaysWhy(t *testing.T) {
	for id, cov := range Coverages(nil) {
		for _, c := range cov.Imperfect {
			if c.Name == "" {
				t.Errorf("%s: an imperfect card has no name", id)
			}
			if !c.Unreviewed && len(c.Caveats) == 0 {
				t.Errorf("%s: %s is counted as a caveat card with no caveat text", id, c.Name)
			}
			for _, text := range c.Caveats {
				if strings.TrimSpace(text) == "" {
					t.Errorf("%s: %s has an empty caveat string", id, c.Name)
				}
			}
		}
	}
}

// TestCoverageProfileSnapshot prints what a player is told about each
// deck today.
//
// It asserts nothing on purpose, and that is worth defending rather
// than apologising for. Any threshold written here ("at least 80%
// full") would be an editorial line drawn by whoever last touched
// this file, enforced on every card PR by a package that owns no
// cards — and it would fail a PR that graded ten previously
// UNREVIEWED cards as `caveats`, which is an improvement to the
// catalog's honesty and not a regression in the decks. The assertions
// that belong here are the ones above: the arithmetic closes, nothing
// is unregistered, and every caveat says why.
//
// What this gives instead is a number nobody has to guess at:
//
//	go test ./internal/decks/ -run Snapshot -v
//
// prints each deck's profile and every imperfect card's reason. Run it
// when you want to know what the picker currently claims — including
// before writing it into a PR body.
func TestCoverageProfileSnapshot(t *testing.T) {
	for _, d := range All() {
		cov := CoverageOf(nil, d)
		t.Logf("%s (%s): %s", d.ID, d.Name, summarise(cov))
		for _, c := range cov.Imperfect {
			if c.Unreviewed && len(c.Caveats) == 0 {
				t.Logf("    %s: (not yet reviewed)", c.Name)
				continue
			}
			t.Logf("    %s: %s", c.Name, strings.Join(c.Caveats, " / "))
		}
	}
}

func summarise(cov Coverage) string {
	if cov.Complete() {
		return fmt.Sprintf("%d cards + %d basic land rows, all fully implemented", cov.Cards, cov.Basics)
	}
	parts := make([]string, 0, 3)
	parts = append(parts, fmt.Sprintf("%d full", cov.Full))
	if cov.Caveats > 0 {
		parts = append(parts, fmt.Sprintf("%d with caveats", cov.Caveats))
	}
	if cov.Unreviewed > 0 {
		parts = append(parts, fmt.Sprintf("%d unreviewed", cov.Unreviewed))
	}
	if cov.Unregistered > 0 {
		parts = append(parts, fmt.Sprintf("%d UNREGISTERED", cov.Unregistered))
	}
	return fmt.Sprintf("%d cards + %d basic land rows — %s", cov.Cards, cov.Basics, strings.Join(parts, ", "))
}
