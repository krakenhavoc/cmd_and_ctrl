package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Everything Comes to Dust — Sorcery {7}{W}{W}{W} (Doctor Who):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you tap
//	 while casting this spell pays for {1} or one mana of that creature's
//	 color.)
//	 Exile all creatures except those that share a creature type with a creature
//	 that convoked this spell, all artifacts, and all enchantments."
//
// Convoke is Spec.TapCost. The creatures that convoked are the ones the cast
// tapped (StackItem.Paid.Convoked); a creature is spared when it shares a
// creature type with any of them (game.SharesCreatureType, a changeling with
// every creature that has a type, CR 702.73a), read as the spell resolves. A
// convoker that has since left the battlefield is read as the card it now is
// (its last types on the battlefield are not kept for this). Every artifact
// and enchantment is exiled, a creature that is one included, and a convoker
// shares a type with itself so it is spared unless it is an artifact or an
// enchantment too. Exile is not destruction: indestructible does not help.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39acc3d0-85fc-4814-93f2-e709c3b2c86b",
		Name:         "Everything Comes to Dust",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepExile, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var convokers []game.Card
			for _, id := range item.Paid.Convoked {
				if c, ok := g.LookupCardForEffect(id); ok {
					convokers = append(convokers, c)
				}
			}
			spared := func(c game.Card) bool {
				for i := range convokers {
					if game.SharesCreatureType(&c, &convokers[i]) {
						return true
					}
				}
				return false
			}
			return ExileAllMatching{Match: func(g *game.Game, caster uuid.UUID, c game.Card) bool {
				if c.IsArtifact() || c.IsEnchantment() {
					return true
				}
				return c.IsCreature() && !spared(c)
			}}.Apply(ctx)
		},
	})
}
