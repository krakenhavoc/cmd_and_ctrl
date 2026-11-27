package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Over Matter — Enchantment {2}{U}{U}{U}{U} (Exodus, S82 OLD b737, #5626):
//
//	"Discard a card: You may tap or untap target artifact, creature, or land."
//
// Rimewind Taskmage's tap-or-untap on a target of three types, paid by a
// discard and nothing else. "You may" is covered by the choice itself:
// tapping a tapped permanent and untapping an untapped one do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "656afde2-cdf3-4907-a0fc-7ac93f5d3e03",
		Name:         "Mind Over Matter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: You may tap or untap target artifact, creature, or land.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.DiscardACard(),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:  effects.B38TapOrUntapTarget("Mind Over Matter"),
		}},
	})
}
