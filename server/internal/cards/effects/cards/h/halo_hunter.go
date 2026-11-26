package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halo Hunter — Creature — Demon {2}{B}{B}{B}, 6/3 (Zendikar, S82 OLD b495):
//
//	"Intimidate
//	 When this creature enters, destroy target Angel."
//
// Rustspore Ram's destroy-on-enter (DestroyFirstLegalCardTarget) aimed at an
// Angel; any Angel is a legal target, the Hunter controller's own included, and
// with none the trigger is not put on the stack. Intimidate is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "293560ec-accf-49c6-9bf4-6d37f56a4eb7",
		Name:            "Halo Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Halo Hunter — destroy target Angel", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target Angel", effects.HasSubtype("Angel"))),
		},
	})
}
