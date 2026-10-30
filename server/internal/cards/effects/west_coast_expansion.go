package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// West Coast Expansion — Sorcery {X}{U}{U} (Marvel Super Heroes Commander,
// #4455):
//
//	"Draw X cards. If X is 5 or more, you may cast a Hero spell from your
//	 hand without paying its mana cost."
//
// The draw is the X spell's (Illuminate). The free cast is the shared
// hand-cast offer (Cecily, Haunted Mage; free_hand_cast.go), made during
// the resolution after the draw (CR 608.2g), so a Hero just drawn is a
// candidate. Only cards with a Hero face that is a spell are offered, and
// with X below 5, or no such card, nothing is asked. The cast is "a Hero
// spell", judged on the card's types: a Hero is a creature or an
// artifact creature card, never a land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e1bf860-22b6-4d08-9b9f-88038cd1dd3f",
		Name:         "West Coast Expansion",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			if ctx.X() < 5 {
				return nil
			}
			controller := ctx.Controller()
			var heroes []uuid.UUID
			for _, id := range allHandCardIDs(ctx.Game, controller) {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && hasSubtype(c, ctx.TextSubtype("Hero")) {
					heroes = append(heroes, id)
				}
			}
			offerFreeHandCastOver(ctx.Game, controller, ctx.Source(), westCoastExpansionOffer,
				"West Coast Expansion — cast a Hero spell from your hand without paying its mana cost?", anySpell, heroes)
			return nil
		},
	})
}

var westCoastExpansionOffer = registerFreeHandCastOffer("choose-cards/west-coast-expansion-free-cast", "West Coast Expansion")
