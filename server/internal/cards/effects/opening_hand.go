package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opening_hand.go — Spec.OpeningHand (CR 103.6, ADR 0133): the actions
// a card offers from its owner's opening hand. The engine asks when
// the mulligan window closes (game/opening_hand.go); a card file only
// declares what it offers.

// BeginTheGameOnTheBattlefield is "If this card is in your opening
// hand, you may begin the game with it on the battlefield" (CR 103.6a)
// — every Leyline, Leyline Axe:
//
//	OpeningHand: BeginTheGameOnTheBattlefield(),
//
// The options are the riders some cards print beside it, each a
// constructor below: Gemstone Caverns' "and you're not the starting
// player" (NotTheStartingPlayer), "with a luck counter on it"
// (WithCounter) and "if you do, exile a card from your hand"
// (ThenExileFromHand).
func BeginTheGameOnTheBattlefield(opts ...OpeningHandOption) *game.OpeningHandAction {
	a := &game.OpeningHandAction{}
	for _, o := range opts {
		o(a)
	}
	return a
}

// OpeningHandOption is one rider on an opening-hand action.
type OpeningHandOption func(*game.OpeningHandAction)

// NotTheStartingPlayer is "and you're not the starting player": the
// action is not offered to the seat that takes the first turn.
func NotTheStartingPlayer() OpeningHandOption {
	return func(a *game.OpeningHandAction) { a.NotStartingPlayer = true }
}

// WithCounter is "with a <kind> counter on it": the card has n counters
// of that kind as the game begins.
func WithCounter(kind string, n int) OpeningHandOption {
	return func(a *game.OpeningHandAction) {
		if a.EntersWithCounters == nil {
			a.EntersWithCounters = map[string]int{}
		}
		a.EntersWithCounters[kind] += n
	}
}

// ThenExileFromHand is "if you do, exile a card from your hand" (n
// cards): their owner chooses them from what is left in their hand.
func ThenExileFromHand(n int) OpeningHandOption {
	return func(a *game.OpeningHandAction) { a.ExileFromHand = n }
}

// checkOpeningHand refuses a declaration the engine would read wrongly,
// at boot: a counter with no name or no count, and a negative exile
// count.
func checkOpeningHand(name string, a *game.OpeningHandAction) {
	if a == nil {
		return
	}
	for kind, n := range a.EntersWithCounters {
		if kind == "" || n <= 0 {
			panic(fmt.Sprintf("effects.Register: %q opening-hand action puts %d %q counters on the card — it needs a named kind and at least one", name, n, kind))
		}
	}
	if a.ExileFromHand < 0 {
		panic(fmt.Sprintf("effects.Register: %q opening-hand action exiles %d cards from the hand", name, a.ExileFromHand))
	}
}
