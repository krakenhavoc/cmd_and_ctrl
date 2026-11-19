package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call from the Grave — Sorcery {2}{B} (Mystery Booster 2, #5948):
//
//	"Put a random creature from a random graveyard into play under your
//	 control. Call from the Grave deals to you an amount of damage equal to
//	 that creature's casting cost."
//
// Two random picks on the keyed random stream: a graveyard among those that
// hold a creature card (a graveyard with none can't give one), then a
// creature card in it. It enters under the caster's control whoever owned it;
// the damage is its mana value (the casting cost of the playtest wording),
// read in the graveyard before it moves, and is dealt to the caster by the
// spell. With no creature card in any graveyard nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "252ff28e-3ea7-4e2c-a30c-3cb80aa278d0",
		Name:         "Call from the Grave",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var owners []uuid.UUID
			for _, p := range g.Seats {
				if p != nil && len(graveyardIDs(g, p.ID, game.Card.IsCreature)) > 0 {
					owners = append(owners, p.ID)
				}
			}
			pickedOwner := g.ChooseAtRandomForEffect(randomDraw(ctx), owners, 1)
			if len(pickedOwner) == 0 {
				return nil
			}
			pickedCard := g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, pickedOwner[0], game.Card.IsCreature), 1)
			if len(pickedCard) == 0 {
				return nil
			}
			card, ok := g.LookupCardForEffect(pickedCard[0])
			if !ok {
				return nil
			}
			mv := card.ManaValue()
			if err := (ReturnFromGraveyard{Target: pickedCard[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
			return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: mv}.Apply(ctx)
		},
	})
}
