package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sephiroth, Planet's Heir — Legendary Creature — Human Avatar Soldier
// {4}{U}{B}, 4/4 (Final Fantasy, #4462):
//
//	"Vigilance
//	 When Sephiroth enters, creatures your opponents control get -2/-2
//	 until end of turn.
//	 Whenever a creature an opponent controls dies, put a +1/+1 counter
//	 on Sephiroth."
//
// The enters trigger is Doomwake Giant's shrink: it applies to the
// creatures opponents control as it resolves (CR 611.2c), not to ones
// that arrive later. The death trigger is Vincent Valentine's condition,
// read by the dead creature's last controller, so a 2-toughness board
// dying to the shrink grows Sephiroth once per creature. The counter goes
// only on the Sephiroth the trigger came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6af71c5a-2c1b-4120-b3c9-c3c70d94a5c9",
		Name:            "Sephiroth, Planet's Heir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, OpponentsOnly: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sephiroth, Planet's Heir — creatures your opponents control get -2/-2 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.OpponentControls()),
					Power:     -2,
					Toughness: -2,
					Label:     "Sephiroth, Planet's Heir — -2/-2 until end of turn",
				})),
			effects.On(game.EventLTB, effects.B33OpponentsCreatureDiedWhen,
				"Sephiroth, Planet's Heir — put a +1/+1 counter on Sephiroth",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
