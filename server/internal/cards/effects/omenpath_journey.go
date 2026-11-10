package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omenpath Journey — Enchantment {3}{G} (The Big Score, #4468):
//
//	"When this enchantment enters, search your library for up to five land
//	 cards that have different names, exile them, then shuffle.
//	 At the beginning of your end step, choose a card at random exiled with
//	 this enchantment and put it onto the battlefield tapped."
//
// The search is Roads Go Ever, Ever On's shape (exile, linked to this
// object, CR 607.2a) with the "different names" set rule the tutors share
// (searchRuleDifferentNames). The end-step trigger reads the linked cards
// when it resolves and picks one on the engine's keyed random stream; with
// none left it does nothing, and an Omenpath that left and came back is a
// new object with no link (CR 400.7).
//
// No simplification.
func init() {
	const name = "Omenpath Journey"
	Register(Spec{
		OracleID:     "8b067e44-2001-4b49-aa4a-2c86f8929751",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 5},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(name+" — search for up to five lands with different names, exile them, then shuffle",
				Do(Instruction(func(ctx *Context) error {
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					return SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return c.IsLand() },
						Dest:      game.ZoneExile,
						Limit:     5,
						Shuffle:   true,
						LinkTo:    game.PermissionCardRef(ref),
						Rule:      searchRuleDifferentNames,
						Reason:    name + " — up to five land cards that have different names",
					}.Apply(ctx)
				}))),
			AtYourEndStep(name+" — put a random exiled land onto the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				var ids []uuid.UUID
				for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
					ids = append(ids, c.InstanceID)
				}
				if len(ids) == 0 {
					return nil
				}
				picked := randomPick(ctx, ids, 1)
				if len(picked) == 0 {
					return nil
				}
				_, err := g.ReturnFromExileToBattlefieldForEffect(picked[0], item.Controller, true)
				return err
			}),
		},
	})
}
