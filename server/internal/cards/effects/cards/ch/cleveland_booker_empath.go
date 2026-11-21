package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleveland Booker, Empath — Legendary Creature — Kwejian Rogue
// {2}{R/G}{R/G}, 2/2 (Star Trek, #4973):
//
//	"Landfall — Whenever a land you control enters, create a 2/2 red and
//	 green Worm creature token with "This token can't block."
//	 {8}: Creatures you control get +2/+0 until end of turn."
//
// One landfall trigger per land (CR 603.2) making the printed Worm (a
// catalog token template whose can't-block line is its own static). The
// pump is Borrowed Grace's first mode as an activated ability: it affects
// the creatures you control as it resolves (CR 611.2c), not ones that
// arrive later this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b80879d3-5abd-47e5-8b8c-61db120854e4",
		Name:         "Cleveland Booker, Empath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Cleveland Booker, Empath — create a 2/2 Worm that can't block (landfall)", game.LandfallPayoff{Tokens: 1},
				effects.Do(effects.CreateToken{Template: effects.TrekB23CantBlockWormToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{8}: Creatures you control get +2/+0 until end of turn.",
			// Instant speed: a team pump answers a block or a burn spell.
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 2}, Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{8}"),
			Effect: effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2,
				Label: "Cleveland Booker, Empath — creatures you control get +2/+0"}),
		}},
	})
}
