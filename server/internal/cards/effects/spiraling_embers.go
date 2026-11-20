package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "25a2f4c8-cdab-49a5-80fd-a75230829f3f",
		Name:         "Spiraling Embers",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return sourceDealsComputedDamageToEachLegalTarget(sokB1CardsInYourHand)(ctx.Game, item)
		},
	})
}
