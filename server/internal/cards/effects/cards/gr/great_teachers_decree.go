package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Teacher's Decree — Sorcery {3}{W}:
//
//	"Creatures you control get +2/+1 until end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The affected set is locked as the spell resolves (CR 611.2c), so a
// creature that enters afterwards is not pumped. Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a6e5834-a89c-44af-b297-1b8ccbfbc907",
		Name:            "Great Teacher's Decree",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.YouControl()),
				Power:     2,
				Toughness: 1,
				Label:     "Great Teacher's Decree — creatures you control get +2/+1",
			}.Apply(ctx)
		},
	})
}
