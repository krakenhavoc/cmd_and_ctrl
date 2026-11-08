package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindlespark Duo — Creature — Lizard Otter {2}{R}, 1/3 (Bloomburrow,
// #4467):
//
//	"{T}: This creature deals 1 damage to target opponent.
//	 Whenever you cast a noncreature spell, untap this creature."
//
// Tal Shiar Analyst's pinger with one target: the tap ability is a CR
// 602 activation with a tap cost, so summoning sickness applies (CR
// 302.6), and the opponent is chosen as it is activated. The untap is
// the Analyst's cast trigger on any noncreature spell its controller
// casts, which resolves above the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "980cd7d5-a511-49ab-bc52-08b71c65db76",
		Name:         "Kindlespark Duo",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target opponent.",
			Cost:    TapCost(),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  DamageToFirstTarget(1),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, b10NoncreatureSpellCastByYou, "Kindlespark Duo — untap", b31UntapSelf),
		},
	})
}
