package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meneldor, Swift Savior — Legendary Creature — Bird Soldier {3}{U}, 3/3 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Flying
//	 Whenever Meneldor deals combat damage to a player, exile up to one
//	 target creature you own, then return it to the battlefield under your
//	 control."
//
// Flying is printed. The trigger is The Neutrinos' "up to one target
// creature you own" clause with a plain blink (Flicker): the creature comes
// back as a new object under its owner's control, so counters and damage
// fall off and its enters triggers fire again. It may target Meneldor
// himself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef9a5bd1-6ce2-4570-9d42-89d997e2fc6f",
		Name:            "Meneldor, Swift Savior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer("Meneldor, Swift Savior — exile up to one target creature you own, then return it to the battlefield under your control",
					effects.FlickerFirstLegalTarget),
				effects.TargetCreature("up to one target creature you own", effects.YouOwn()).WithCount(0, 1)),
		},
	})
}
