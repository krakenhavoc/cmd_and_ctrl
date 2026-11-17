package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ertai, the Corrupted — Legendary Creature — Phyrexian Human Wizard
// {2}{W}{U}{B}, 3/4 (Planeshift, #5615):
//
//	"{U}, {T}, Sacrifice a creature or enchantment: Counter target
//	 spell."
//
// Vile Mutilator's "a creature or enchantment" as a sacrifice clause
// beside the mana and the tap, and Tidal Control's counterTheChosenSpell.
// Ertai is a creature, so he can be his own sacrifice. The tap means
// summoning sickness applies (CR 302.6). A spell that has left the stack
// by the time the ability resolves is no longer a legal target, and the
// ability does nothing (CR 608.2b); the sacrifice is spent all the same.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36934bd0-b275-4222-926c-b5a74cf0967d",
		Name:         "Ertai, the Corrupted",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}, Sacrifice a creature or enchantment: Counter target spell.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{U}"), TapCost(), SacrificeN(1, "a creature or enchantment", Or(Creature(), Enchantment()))),
			Targets: TargetSpell("target spell"),
			Effect:  counterTheChosenSpell,
		}},
	})
}
