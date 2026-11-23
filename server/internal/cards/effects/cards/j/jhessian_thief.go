package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhessian Thief — Creature — Human Rogue {2}{U}, 1/3 (Magic Origins, #5549):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// Prowess is the engine keyword; the draw is Scroll Thief's combat-damage
// trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b917c3aa-2d54-49e0-924a-dfd02eaf4b2d",
		Name:            "Jhessian Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Jhessian Thief — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
