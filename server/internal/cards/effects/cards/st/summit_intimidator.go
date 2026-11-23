package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summit Intimidator — Creature — Yeti {3}{R}, 4/3 (Tarkir: Dragonstorm,
// #4463):
//
//	"Reach
//	 When this creature enters, target creature can't block this turn."
//
// Zog, Triceraton Castaway's entry trigger: any creature may be the
// target, the Intimidator included, and the restriction lasts until the
// end of the turn.
//
// No simplification.
func init() {
	const label = "Summit Intimidator — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:        "b772d35a-d365-47ce-b1dd-a2ba574c74ea",
		Name:            "Summit Intimidator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
					effects.TargetCreature("target creature")),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
