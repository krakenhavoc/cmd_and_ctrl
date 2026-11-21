package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiraling Embers — Sorcery — Arcane {3}{R} (Saviors of Kamigawa, #5601):
//
//	"Spiraling Embers deals damage to any target equal to the number of
//	 cards in your hand."
//
// The hand is counted as the spell resolves (CR 608.2h), with Spiraling
// Embers already off it; a target that left in response is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25a2f4c8-cdab-49a5-80fd-a75230829f3f",
		Name:         "Spiraling Embers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SourceDealsComputedDamageToEachLegalTarget(effects.SokB1CardsInYourHand)(ctx.Game, item)
		},
	})
}
