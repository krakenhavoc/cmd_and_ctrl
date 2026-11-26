package cardsth

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Thassa's Ire — Enchantment {U} (Journey into Nyx, S82 OLD, #5558):
//
//	"{3}{U}: You may tap or untap target creature."
//
// Bounding Krasis's tap-or-untap choice on an activated ability: the creature
// is targeted as the ability goes on the stack and its controller picks tap,
// untap or neither as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb3b365b-eb8c-4695-bfd6-88060ad20741",
		Name:         "Thassa's Ire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: You may tap or untap target creature.",
			Cost:    effects.ManaCost("{3}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NemMayTapOrUntapTheTarget("Thassa's Ire"),
		}},
	})
}
