package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feast of Sanity — Enchantment {3}{B} (Modern Horizons 2):
//
//	"Whenever you discard a card, this enchantment deals 1 damage to any target
//	 and you gain 1 life."
//
// One trigger per card you discard, the cleanup discard included; the target is
// chosen as it goes on the stack and the Feast deals the damage as it last
// existed if it has left. The life is gained whatever the damage did (Balefire
// Liege's damage-then-life body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af384946-e22a-45cd-ba07-1a5fbd4a3f0c",
		Name:         "Feast of Sanity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return effects.DiscardedByYou(ev, source)
					}, "Feast of Sanity — 1 damage to any target and you gain 1 life", effects.GapB106OneDamageAndOneLife),
					effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
