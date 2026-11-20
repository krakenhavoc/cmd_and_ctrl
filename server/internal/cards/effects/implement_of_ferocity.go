package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Implement of Ferocity — Artifact {1} (Aether Revolt,
// #5536):
//
//	"{G}, Sacrifice this artifact: Put a +1/+1 counter on target creature.
//	 Activate only as a sorcery.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// Greta, Sweettooth Scourge's sorcery-speed counter row, paid with
// the Implement itself.
// The draw is Chromatic Star's leaves-for-the-graveyard trigger: it
// fires however the Implement got there, sacrificed to its own ability
// or destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b92629c0-1619-4214-929f-4404992fabeb",
		Name:         "Implement of Ferocity",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{G}, Sacrifice this artifact: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{G}"), SacrificeThis()),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       putACounterOnTheTarget(game.CounterPlusOne),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Implement of Ferocity — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
