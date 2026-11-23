package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Mare — Creature — Nightmare Horse {1}{B}{B}, 2/2
// (Core Set 2019, #5527):
//
//	"This creature can't be blocked by white creatures.
//	When this creature enters, creatures your opponents control get -1/-1 until end of turn."
//
// Dauthi Horror's block rule for white creatures, and Doomwake Giant's
// -1/-1 sweep of the opponents' creatures as an enters trigger. The
// set is locked when it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cd096ac-dd67-4208-a3ad-8e6061b442c6",
		Name:         "Plague Mare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.OfColor("W"), "white creatures"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Plague Mare — creatures your opponents control get -1/-1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.OpponentControls()),
					Power:     -1,
					Toughness: -1,
					Label:     "Plague Mare — -1/-1 until end of turn",
				})),
		},
	})
}
