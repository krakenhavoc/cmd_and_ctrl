package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherspouts — Instant {3}{U}{U} (EDHREC rank 2780):
//
//	"For each attacking creature, its owner puts it on their choice of
//	 the top or bottom of their library."
//
// Aetherize's big sibling: the attackers do not come back to hand,
// they go into the library, and each owner decides where. Every
// attacker goes onto its owner's library together (one simultaneous
// exit), and each owner is then asked a put_in_library with the
// top_or_bottom placement (ADR 0088) over their own: each card to the
// top or the bottom, and the order of each pile (CR 401.4 hands the
// owner that order anyway). Owners are asked in seat order; nobody
// without an attacker is asked.
//
// #996: the choice used to be asked as a SCRY over the tucked cards.
// That is a keyword action the card never prints — it fired every
// "whenever you scry" payoff and opened the CR 614 scry window, so a
// scry-count replacement reached past the attackers into the library.
//
// The set is every attacking creature, any controller — a creature
// the caster controls that was goaded into attacking goes too, as
// printed.
//
// One engine-wide gap, not this card's: a token put into a library
// (or a hand, by any bounce) is not swept out of it as CR 111.8
// asks — it sits there as a card. Reported on #388; Evacuation and
// Aetherize share it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48369aec-a991-4bef-8554-01c84302b063",
		Name:         "Aetherspouts",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b26TuckAttackersTopOrBottomByOwnersChoice(ctx)
		},
	})
}
