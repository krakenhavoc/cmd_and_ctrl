package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Inertia — Enchantment — Aura Curse {2}{U} (Commander 2013, S82
// older sets):
//
//	"Enchant player
//	 Whenever a player attacks enchanted player with one or more creatures,
//	 that attacking player may tap or untap target permanent of their
//	 choice."
//
// One trigger per attack declaration per attacking player, controlled by that
// player (Curse of Opulence's batching), so the target and the "may" are
// theirs. The target is judged again as the ability resolves (CR 608.2b); the
// tap or untap is the shared "may tap or untap" prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bbeb0ee-647b-43d3-91b3-6869d5ccb8b8",
		Name:         "Curse of Inertia",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OldB383WhenAPlayerAttacksEnchantedPlayer("Curse of Inertia — the attacking player may tap or untap target permanent",
					effects.NemMayTapOrUntapTheTarget("Curse of Inertia")),
				effects.OfThatPlayersChoice(effects.TargetPermanent("target permanent of their choice"))),
		},
	})
}
