package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Astrid Peth — Legendary Creature — Human {1}{W}, 2/2 (Doctor Who, #5948):
//
//	"Whenever Astrid Peth enters or attacks, create a Food token.
//	 Brand-new Sky — Whenever you sacrifice a Clue or Food, Astrid Peth
//	 explores."
//
// The first ability is one trigger on two events (WhenThisEntersOrAttacks).
// The second reads a sacrifice by the controller of a Clue or Food as the
// permanent leaves (Eloise's reading: the type is read while it is still on
// the battlefield); Astrid explores each time (CR 701.44). If Astrid has left
// she still reveals the card and gets no counter (CR 701.44c). "Brand-new
// Sky" is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa55177c-ba65-4113-9c05-2d63318cf7c5",
		Name:         "Astrid Peth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Astrid Peth — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.HasSubtype("Clue") || c.HasSubtype("Food"))
			}, "Astrid Peth — she explores", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Explores{Explorer: ctx.Source()}.Apply(ctx)
			}),
		},
	})
}
