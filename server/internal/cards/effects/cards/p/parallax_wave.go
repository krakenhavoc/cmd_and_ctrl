package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parallax Wave — Enchantment {2}{W}{W} (Nemesis, S82 OLD, #5618):
//
//	"Fading 5
//	 Remove a fade counter from this enchantment: Exile target creature.
//	 When this enchantment leaves the battlefield, each player returns to the
//	 battlefield all cards they own exiled with it."
//
// Parallax Tide's shape with a creature as the prey: fading is the shared pair
// (fading.go), the creature is exiled linked to this object, and the leave
// trigger returns every card still exiled with it together, under its owner's
// control. A token exiled this way is gone for good, and a Wave that returns is
// a new object with no link.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0accfb04-bc8b-4cab-997a-2337ac42db8f",
		Name:         "Parallax Wave",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Fading(5)},
		Triggered: []game.TriggeredAbility{
			effects.FadingUpkeep(),
			effects.On(game.EventLTB, effects.Self, "Parallax Wave — each player returns the cards they own exiled with it", effects.ReturnAllExiledWithThisTogether),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a fade counter from this enchantment: Exile target creature.",
			Cost:    effects.RemoveCountersFromThis(game.CounterFade, 1),
			Targets: effects.TargetCreature("target creature"),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.AvB7ExileTargetLinked,
		}},
	})
}
