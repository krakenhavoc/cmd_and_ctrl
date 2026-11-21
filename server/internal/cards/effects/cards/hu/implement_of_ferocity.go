package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b92629c0-1619-4214-929f-4404992fabeb",
		Name:         "Implement of Ferocity",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{G}, Sacrifice this artifact: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{G}"), effects.SacrificeThis()),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       effects.PutACounterOnTheTarget(game.CounterPlusOne),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Implement of Ferocity — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
