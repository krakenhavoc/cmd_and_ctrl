package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Night Clubber — Creature — Human Warrior {1}{B}{B}, 2/2:
//
//	"When this creature enters, creatures your opponents control get
//	 -1/-1 until end of turn.
//	 Blitz {2}{B}"
//
// Blitz (CR 702.152, #3772). The shrink is BoostUntilEOT over a Match,
// so the affected set is locked as the trigger resolves (CR 611.2c): a
// creature an opponent casts afterwards is not shrunk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4d2abfe1-f5f8-4ed5-915d-bfe98c6321ee",
		Name:             "Night Clubber",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Blitz("{2}{B}")},
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Night Clubber — creatures your opponents control get -1/-1",
				effects.Do(effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.OpponentControls()),
					Power:     -1,
					Toughness: -1,
					Label:     "Night Clubber — creatures your opponents control get -1/-1",
				})),
		},
	})
}
