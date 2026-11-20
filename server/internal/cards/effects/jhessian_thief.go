package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b917c3aa-2d54-49e0-924a-dfd02eaf4b2d",
		Name:            "Jhessian Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Jhessian Thief — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
