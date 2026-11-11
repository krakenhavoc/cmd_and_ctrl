package deck

import (
	"fmt"
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// limited.go is the deck check of limited play (ADR 0222 §2, #5440): a
// deck built from the cards one player drafted, plus basic lands.
//
// The rules it enforces, from the Comprehensive Rules:
//
//	CR 100.2b: "In limited play (a way of playing in which each player
//	gets the same quantity of unopened Magic product such as booster
//	packs and creates their own deck using only this product and basic
//	land cards), each deck has a minimum deck size of 40 cards. A
//	limited deck may contain as many duplicates of a card as are
//	included with the product."
//
//	CR 100.4b: "In limited play involving individual players, all cards
//	in a player's card pool not included in their deck are in that
//	player's sideboard."
//
// and CR 905.1d, which says the same of a draft: "The player builds
// their deck from only these cards and any number of basic land cards."
//
// So a deck is legal when it has at least 40 cards, every card that is
// not a basic land is in the pool, and it holds no more copies of one
// than the pool does. A basic land is free: any number, whether or not
// the pool holds one. There is no four-card limit and no legality map;
// the pool is the format.

// LimitedCard is one card as the limited check reads it: its English
// name, and whether it is a basic land.
type LimitedCard struct {
	Name  string
	Basic bool
}

// Violation codes of the limited check.
const (
	// CodeNotInPool flags a card the player did not draft.
	CodeNotInPool = "not_in_pool"
	// CodeMoreThanPool flags more copies of a card than the pool holds.
	CodeMoreThanPool = "more_copies_than_pool"
)

// ValidateLimited checks a limited deck against the pool it was built
// from. The error is a *ValidationError, one violation per problem, in
// words for the player.
func ValidateLimited(main, pool []LimitedCard) error {
	spec := game.FormatLimited.Spec()
	var vs []Violation
	if n := len(main); n < spec.Deck.MinCards {
		vs = append(vs, Violation{
			Code:    CodeWrongCardCount,
			Message: fmt.Sprintf("deck has %d cards; a %s deck needs at least %d", n, spec.Name, spec.Deck.MinCards),
		})
	}
	have := map[string]int{}
	for _, c := range pool {
		have[c.Name]++
	}
	want := map[string]int{}
	var names []string
	for _, c := range main {
		if c.Basic {
			// Any number of basic lands (CR 100.2b).
			continue
		}
		if want[c.Name] == 0 {
			names = append(names, c.Name)
		}
		want[c.Name]++
	}
	sort.Strings(names)
	for _, name := range names {
		switch n, held := want[name], have[name]; {
		case held == 0:
			vs = append(vs, Violation{
				Code:    CodeNotInPool,
				Card:    name,
				Message: fmt.Sprintf("%q is not in your card pool; a %s deck is built from the cards you drafted and basic lands", name, spec.Name),
			})
		case n > held:
			vs = append(vs, Violation{
				Code:    CodeMoreThanPool,
				Card:    name,
				Message: fmt.Sprintf("%q appears %d times; your card pool holds %d", name, n, held),
			})
		}
	}
	if len(vs) == 0 {
		return nil
	}
	return &ValidationError{Violations: vs}
}

// LimitedSideboard is the sideboard of a limited deck: every card of
// the pool the deck does not hold (CR 100.4b). A basic land the deck
// added from outside the pool takes nothing from it. It returns the
// indexes into pool, in pool order.
func LimitedSideboard(main, pool []LimitedCard) []int {
	used := map[string]int{}
	for _, c := range main {
		used[c.Name]++
	}
	var side []int
	for i, c := range pool {
		if used[c.Name] > 0 {
			used[c.Name]--
			continue
		}
		side = append(side, i)
	}
	return side
}
