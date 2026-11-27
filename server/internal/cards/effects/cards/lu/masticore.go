package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Masticore — {4} Artifact Creature — Masticore 4/4 (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you
//	 discard a card.
//	 {2}: This creature deals 1 damage to target creature.
//	 {2}: Regenerate this creature."
//
// Molten-Tail Masticore's upkeep discard (OldB472UpkeepSacrificeUnlessDiscard)
// and regeneration row; the damage ability targets a creature and is dealt by
// this creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "621476a4-44bf-4f55-adb1-880cdd213f38",
		Name:         "Masticore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Masticore — sacrifice this creature unless you discard a card",
				effects.OldB472UpkeepSacrificeUnlessDiscard("Masticore")),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: This creature deals 1 damage to target creature.",
				Cost:    effects.ManaCost("{2}"),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.B33DamageChosenTargetFromSource(1),
			},
			effects.LrwB1RegenerateThisRow("{2}"),
		},
	})
}
