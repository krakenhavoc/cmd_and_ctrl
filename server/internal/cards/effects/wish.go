package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wish — Sorcery {2}{R}:
//
//	"You may play a card you own from outside the game this turn."
//
// PlayFromOutside (#3955, ADR 0198's 2026-10-16 amendment): when Wish
// resolves the owner chooses a card from outside the game (their
// sideboard, or the card database when the deck seated none) and may
// play it until end of turn, at its own timing and for its own cost. A
// land counts: the permission is not cast-only, so playing it spends the
// turn's land drop. A card not played stays outside the game.
//
// The card is NAMED when Wish resolves rather than at the moment it is
// played. A database card has no instance until it is named (ADR 0198
// section 2), so the choice cannot be deferred there; it is made the
// same way for a sideboard so the card behaves the same in every deck.
// ADR 0198's 2026-10-16 amendment records the posture.
func init() {
	Register(Spec{
		OracleID:     "2ccbf0ee-c856-4f67-8579-811fe8571a2e",
		Name:         "Wish",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PlayFromOutside{
				Name:   "Wish",
				Clause: "a card",
				Max:    1,
			}.Apply(ctx)
		},
	})
}
