package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lashknife Barrier — Enchantment {2}{W} (Planeshift, S82 OLD #5615):
//
//	"When this enchantment enters, draw a card.
//	 If a source would deal damage to a creature you control, it deals that
//	 much damage minus 1 to that creature instead."
//
// The static clause is a replacement of the damage event (CR 614.1a, not a
// prevention), reducing damage to creatures you control by 1; damage to you is
// untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfde73b2-46bd-4fdb-9468-35819af609f7",
		Name:         "Lashknife Barrier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lashknife Barrier — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Replacements: []game.ReplacementEffect{effects.OldB637DamageToYourCreaturesMinusOne("Lashknife Barrier")},
	})
}
