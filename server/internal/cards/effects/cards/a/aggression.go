package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aggression — Enchantment — Aura {2}{R} (Ice Age, S82 OLD #5636):
//
//	"Enchant non-Wall creature
//	 Enchanted creature has first strike and trample.
//	 At the beginning of the end step of enchanted creature's controller, destroy
//	 that creature if it didn't attack this turn."
//
// The per-object attack tally is the turn's (TimesAttackedThisTurn), as Hexhaven
// Dueling Arena reads it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6c38a2b-2fe8-4080-80ce-cfa12aaf57f0",
		Name:         "Aggression",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.NonWall()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("first strike", "trample")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.EndStepOfEnchantedCreaturesControllerB829,
				"Aggression — destroy enchanted creature if it didn't attack this turn", effects.AggressionDestroyB829),
		},
	})
}
