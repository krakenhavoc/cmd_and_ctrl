package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karok Wrangler — Creature — Elf Druid {4}{G}, 3/3 (Strixhaven: School of
// Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, put a
//	 +1/+1 counter on target creature you control."
//
// The shared Magecraft trigger with a target (Silverquill Apprentice's
// shape): a creature you control, chosen as the trigger goes on the stack,
// gets the counter if it is still a legal target as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "031f629f-7a9a-4143-9f5d-434ed1c23b7f",
		Name:         "Karok Wrangler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Magecraft("Karok Wrangler — put a +1/+1 counter on target creature you control (magecraft)",
					effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
