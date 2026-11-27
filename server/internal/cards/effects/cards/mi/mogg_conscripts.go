package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Conscripts — Creature — Goblin {R}, 2/2 (Tempest, S82 OLD, #5628):
//
//	"This creature can't attack unless you've cast a creature spell this turn."
//
// Goblin Cohort's text through its one shared restriction: an attack-target
// restriction (CR 508.1c) read live at declaration from the per-turn cast tally.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b015686d-07e1-46c7-8e51-2bec646860d6",
		Name:         "Mogg Conscripts",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB595CantAttackUnlessYouHaveCastACreatureSpell()},
	})
}
