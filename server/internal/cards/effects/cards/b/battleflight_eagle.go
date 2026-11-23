package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battleflight Eagle — Creature — Bird {4}{W}, 2/2 (Magic 2013, #5568):
//
//	"Flying
//	 When this creature enters, target creature gets +2/+2 and gains
//	 flying until end of turn."
//
// Arborea Pegasus at +2/+2: the target, the Eagle itself included, gets
// both halves if it is still legal as the trigger resolves (CR 608.2b)
// and neither if it is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f36e316d-c0ad-414d-8cf0-b1c9ea779e03",
		Name:            "Battleflight Eagle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Battleflight Eagle — target creature gets +2/+2 and gains flying until end of turn",
					effects.TrekB41PumpAndGrantEachTarget(2, 2, "Battleflight Eagle — +2/+2 and flying", "flying")),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying"))),
		},
	})
}
