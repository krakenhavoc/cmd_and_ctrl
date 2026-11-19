package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "28274c39-19e2-4b4c-bd53-d31d6063c2ba",
		Name:            "Weaver of Lightning",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, YouCast(Or(Instant(), Sorcery())),
				"Weaver of Lightning — 1 damage to target creature an opponent controls",
				DealDamageToTheTarget(1)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
