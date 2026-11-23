package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Games — Instant {U} (Stronghold, #5627):
//
//	"Buyback {2}{U} (You may pay an additional {2}{U} as you cast this
//	 spell. If you do, put this card into your hand as it resolves.)
//	 Tap target artifact, creature, or land."
//
// Ring of Gix's target clause and Elder Deep-Fiend's shared
// tapEachLegalTarget body, with Capsize's buyback (CR 702.27a). Tapping
// a permanent that is already tapped does nothing (CR 701.26a), and the
// spell still resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ea43372f-b0a7-4d1b-be3e-a4a738b06d59",
		Name:          "Mind Games",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{2}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
