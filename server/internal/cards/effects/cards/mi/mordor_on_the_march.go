package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mordor on the March — Sorcery {3}{B}{R} (Tales of Middle-earth
// Commander):
//
//	"Exile a creature card from your graveyard. Create a token that's a copy
//	 of it. It gains haste until end of turn. Exile it at the beginning of
//	 the next end step.
//	 Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn.)"
//
// Storm is the shared trigger (CR 702.40a), so each copy repeats the whole
// resolution. The creature card is chosen from your graveyard as the spell
// resolves (it is not a target); the exile is a keyed tail, so the token is
// copied from the card as it really reached exile (a card the CR 614 window
// sent elsewhere is not copied) and Mimic Vat's token copy takes over: the
// token has haste (it is exiled at the next end step, so "until end of turn"
// and "while it lasts" are the same) and a delayed trigger exiles it then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60fa9f5f-2360-4d5e-b291-1318803e3c9f",
		Name:         "Mordor on the March",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			cands := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			if len(cands) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    ctx.Controller(),
				FromPlayer: ctx.Controller(),
				Source:     item.SourceCardID,
				Question:   "Mordor on the March — exile a creature card from your graveyard",
				Cards:      cands,
				Min:        1,
				Max:        1,
				LiveZone:   game.ZoneGraveyard,
				ThenKey:    s82GapB120MordorPicked,
				Asked:      game.NewAskedItem(item),
			})
			return nil
		},
	})
}

var s82GapB120MordorPicked = effects.AskedPick("choose-cards/mordor-on-the-march-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(s82GapB120MordorCopy)}.Apply(ctx)
})

var s82GapB120MordorCopy = effects.CardTail("mordor-on-the-march/copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.TrekB40CreateCopyExiledAtNextEndStep(ctx, ctx.Controller(), o.IDs[0], effects.TokenCopyGainsHaste, "Mordor on the March — exile the token")
})
