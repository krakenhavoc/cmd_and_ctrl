package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of the Nightly Hunt — Enchantment — Aura Curse {2}{R} (Commander 2015,
// S82 OLD #5546):
//
//	"Enchant player
//	 Creatures enchanted player controls attack each combat if able."
//
// Warmonger Hellkite's attack requirement (AttacksEachCombatWhere), scoped to
// the creatures the enchanted player controls, read off the Aura's attachment
// as combat begins so it ends when the Curse leaves or the player does. It is
// a requirement of the Curse, so CR 508.1d's "if able" lets a tapped or
// summoning-sick creature stay home.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc6f2b79-8fcd-4542-904b-943c7493b8ad",
		Name:         "Curse of the Nightly Hunt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Static: []game.StaticAbility{
			effects.AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && target.Controller == source.AttachedTo.ID
			}),
		},
	})
}
