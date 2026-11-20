package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hell Mongrel — Creature — Nightmare Dog {3}{B}, 4/3 (Modern Horizons 2,
// #5501):
//
//	"Discard a card: This creature gets +1/+1 until end of turn.
//	 Madness {2}{B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Noose Constrictor's discard pump (the discard is a cost named at
// announce) and the engine's madness (CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64aa732c-d939-48c7-b1d6-c73fa1eedb40",
		Name:         "Hell Mongrel",
		Completeness: CompletenessFull,
		Madness:      "{2}{B}",
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    DiscardN(1, "a card"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				src := ctx.Source()
				if !onBattlefield(g, src) {
					return nil
				}
				return BoostUntilEOT{Target: src, Power: 1, Toughness: 1, Label: "Hell Mongrel — +1/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
