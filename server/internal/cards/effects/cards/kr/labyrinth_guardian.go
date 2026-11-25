package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Labyrinth Guardian — {1}{U} Creature — Illusion Warrior 2/3 (Amonkhet, S82
// old batch 247):
//
//	"When this creature becomes the target of a spell, sacrifice it.
//	 Embalm {3}{U}"
//
// The trigger fires on a SPELL targeting it, not an ability (Dream Strix's
// shape): the sacrifice resolves above the spell, so the spell fizzles when
// the Guardian is its only target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "867ee3e7-88df-4e4b-b199-f441a8460573",
		Name:         "Labyrinth Guardian",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Embalm("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.SelfTargetedByASpell, "Labyrinth Guardian — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
