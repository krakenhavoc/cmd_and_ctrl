package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frantic Firebolt — Instant {2}{R} (Wilds of Eldraine, #4471):
//
//	"Frantic Firebolt deals X damage to target creature, where X is 2 plus
//	 the number of cards in your graveyard that are instant cards, sorcery
//	 cards, and/or have an Adventure."
//
// Broadside Bombardiers' computed damage to the one target
// (sourceDealsComputedDamageToEachLegalTarget), the amount read once as
// the spell resolves. Frantic Firebolt is still on the stack then, so it
// does not count itself; each card in the graveyard counts once however
// many of the three it is (instantSorceryOrAdventureCardsInGraveyard). A
// target that has left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1882cc8-c81a-4131-b639-1bd6e083f8ba",
		Name:         "Frantic Firebolt",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return sourceDealsComputedDamageToEachLegalTarget(func(c *Context) int {
				return 2 + instantSorceryOrAdventureCardsInGraveyard(c.Game, c.Controller())
			})(ctx.Game, item)
		},
	})
}
