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
		// The exile offer waits for a prompt the draws ask (a dredge offer; CR 608.2c, #4155).
		OnResolve: ResolveSteps(DrawCards{N: 2}, Instruction(func(ctx *Context) error {
			var copies []uuid.UUID
			if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Sphinx's Approach" && c.InstanceID != ctx.Item.SourceCardID {
						copies = append(copies, c.InstanceID)
					}
				}
			}
			if len(copies) < 4 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Item.Controller,
				Source:   ctx.Item.SourceCardID,
				Question: "Sphinx's Approach — exile this spell and four cards named Sphinx's Approach from your graveyard (choose exactly four, or none)",
				Cards:    copies,
				Min:      0,
				Max:      4,
				LiveZone: game.ZoneGraveyard,
				ThenKey:  sphinxsApproachExile,
				Asked:    game.NewAskedItem(ctx.Item),
			})
			return nil
		})),
	})
}

// sphinxsApproachExile is the four copies exiled with this spell and then the
// Sphinx search, keyed (#4253 sweep). "Exactly four, or none" is a registered
// set rule, so the restored prompt still refuses any other count.
var sphinxsApproachExile = askedPickRule("choose-cards/sphinxs-approach-exile-four",
	func(picked []game.Card) bool { return len(picked) == 4 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 4 {
			return nil
		}
		for _, id := range r.Picked {
			if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		if !sourceIsNewObject(ctx.Game, ctx.Item) {
			if err := ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID); err != nil {
				return err
			}
		}
		return SearchLibrary{
			Player: ctx.Item.Controller,
			// SearchLibrary lends the spell text (CR 612.2, #3984).
			Predicate: func(card game.Card) bool { return card.IsCreature() && card.HasSubtypeWord("Sphinx") },
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Shuffle:   true,
			Reason:    "Sphinx's Approach — a Sphinx creature card",
			Source:    ctx.Item.SourceCardID,
		}.Apply(ctx)
	})
