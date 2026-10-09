package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sphinx's Approach — Instant {1}{U}{U} (Reality Fracture, tracker #2795):
//
//	"Draw two cards. Then you may exile this spell and four cards named
//	 Sphinx's Approach from your graveyard. If you do, search your
//	 library for a Sphinx creature card, put it onto the battlefield,
//	 then shuffle.
//	 A deck can have any number of cards named Sphinx's Approach."
//
// The "may" is a card pick over the copies in your graveyard that must
// be exactly four or nothing (the set rule on the prompt), so picking
// fewer declines. The spell exiles itself with them, which is why it
// moves itself rather than being left for the stack to bury. The
// search is mandatory once the exile has happened and ends in a
// shuffle. With fewer than four copies in the graveyard no question is
// asked. The any-number deck rule is read from the oracle text by the
// deck check, as Relentless Rats'.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48b1ac27-430f-4fd9-a71b-6ad125c57fe3",
		Name:         "Sphinx's Approach",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
				return err
			}
			var copies []uuid.UUID
			if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Sphinx's Approach" && c.InstanceID != item.SourceCardID {
						copies = append(copies, c.InstanceID)
					}
				}
			}
			if len(copies) < 4 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Sphinx's Approach — exile this spell and four cards named Sphinx's Approach from your graveyard (choose exactly four, or none)",
				Cards:    copies,
				Min:      0,
				Max:      4,
				Zone:     game.ZoneGraveyard,
				Validate: func(picked []game.Card) bool { return len(picked) == 4 },
				Then: func(g *game.Game, picked []uuid.UUID) error {
					if len(picked) != 4 {
						return nil
					}
					c := NewContext(g, item)
					for _, id := range picked {
						if err := (ExileTarget{Target: id}).Apply(c); err != nil {
							return err
						}
					}
					if !sourceIsNewObject(g, item) {
						if err := g.ExileCardForEffect(item.SourceCardID); err != nil {
							return err
						}
					}
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(card game.Card) bool { return card.IsCreature() && card.HasSubtype("Sphinx") },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Sphinx's Approach — a Sphinx creature card",
						Source:    item.SourceCardID,
					}.Apply(c)
				},
			})
			return nil
		},
	})
}
