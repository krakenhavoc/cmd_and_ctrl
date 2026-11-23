package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weaver of Lightning — Creature — Human Shaman {2}{R}, 1/4:
//
//	"Reach
//	 Whenever you cast an instant or sorcery spell, this creature deals 1
//	 damage to target creature an opponent controls."
//
// The trigger targets as it goes on the stack (CR 603.3d); a target that
// is illegal or gone by resolution takes nothing (CR 608.2b), and the
// damage is dealt by the Weaver as last known if it has left (CR 608.2h).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "28274c39-19e2-4b4c-bd53-d31d6063c2ba",
		Name:            "Weaver of Lightning",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery())),
				"Weaver of Lightning — 1 damage to target creature an opponent controls",
				effects.DealDamageToTheTarget(1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
