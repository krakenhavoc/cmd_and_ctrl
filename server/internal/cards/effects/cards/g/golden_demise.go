package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golden Demise — Sorcery {1}{B}{B} (EDHREC rank 18368):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 All creatures get -2/-2 until end of turn. If you have the city's
//	 blessing, instead only creatures your opponents control get -2/-2
//	 until end of turn."
//
// A symmetrical shrink-ray for a go-wide deck that turns into a one-sided
// sweeper once the caster has ten permanents. The blessing is read as the
// spell resolves, after its own ascend check (CR 702.131a), so a caster
// who reaches ten permanents with this very resolution gets the
// one-sided half. The set of creatures is locked as the effect begins
// (CR 611.2c), as BoostUntilEOT does for every "all creatures" boost.
//
// The Purpose declaration is the symmetrical sweep: it is the bot's
// reading of the card as cast by a player without the blessing, and the
// one-sided half is only ever better for the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4648fa3-0343-414b-b04d-07fe0d132145",
		Name:            "Golden Demise",
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			match := effects.Creature()
			label := "Golden Demise — all creatures get -2/-2"
			if effects.YouHaveTheCitysBlessing(ctx.Game, item.Controller) {
				match = effects.And(effects.Creature(), effects.OpponentControls())
				label = "Golden Demise — creatures your opponents control get -2/-2"
			}
			return effects.BoostUntilEOT{Match: match, Power: -2, Toughness: -2, Label: label}.Apply(ctx)
		},
	})
}
