package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ardent Dustspeaker — Creature — Minotaur Shaman {4}{R}, 3/4 (Strixhaven:
// School of Mages, #5502):
//
//	"Whenever this creature attacks, you may put an instant or sorcery card from
//	 your graveyard on the bottom of your library. If you do, exile the top two
//	 cards of your library. You may play those cards this turn."
//
// The pick is optional (zero or one instant or sorcery card of yours in the
// graveyard) and re-checked as the answer arrives; with a card put on the
// bottom ("if you do") the two cards exiled are the top two as the library now
// is, and they may be played, lands included, this turn (CR 118.9a): the
// ordinary impulse exile (ExileTopWithPermission). With no card to put, nothing
// is asked and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "370974e7-d242-4f24-a031-cba71c935306",
		Name:         "Ardent Dustspeaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ardent Dustspeaker — you may put an instant or sorcery card from your graveyard on the bottom of your library; if you do, exile the top two cards and you may play them this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					cands := effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
					if len(cands) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    item.Controller,
						FromPlayer: item.Controller,
						Source:     item.SourceCardID,
						Question:   "Ardent Dustspeaker — you may put an instant or sorcery card from your graveyard on the bottom of your library",
						Cards:      cands,
						Min:        0,
						Max:        1,
						LiveZone:   game.ZoneGraveyard,
						ThenKey:    ardentDustspeakerPicked,
						Asked:      game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// ardentDustspeakerPicked puts the chosen card on the bottom of its owner's
// library and, if it went, exiles the top two cards with a play permission.
// An on-disk key.
var ardentDustspeakerPicked = effects.AskedPick("choose-cards/old-b044-ardent-dustspeaker-picked", func(ctx *effects.Context, r game.CardsChosen) error {
	var id uuid.UUID
	if len(r.Picked) > 0 {
		id = r.Picked[0]
	}
	if id == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !(c.IsInstant() || c.IsSorcery()) || c.Owner != ctx.Controller() {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	if err := ctx.Game.TuckToLibraryForEffect(id, true); err != nil {
		return err
	}
	return effects.ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: 2}.Apply(ctx)
})
