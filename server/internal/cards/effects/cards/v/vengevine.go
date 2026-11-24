package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vengevine — Creature — Elemental {2}{G}{G}, 4/3 (Secret Lair Drop):
//
//	"Haste
//	 Whenever you cast a spell, if it's the second creature spell you cast
//	 this turn, you may return this card from your graveyard to the
//	 battlefield."
//
// A graveyard trigger (InGraveyard) with an optional return. The intervening
// "if" is read off the cast tally as the spell becomes cast: it is a creature
// spell and the creature spells cast this turn (all spells less the
// noncreature ones) number two. It is checked again as the trigger resolves
// by the card still being this object in the graveyard (CR 603.4, 400.7).
func init() {
	const label = "Vengevine — you may return this card from your graveyard to the battlefield"
	effects.Register(effects.Spec{
		OracleID:        "7ed52301-81ea-4e7f-b985-cfab0593cae4",
		Name:            "Vengevine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.Optional(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !spell.IsCreature() {
					return false
				}
				tally := g.CastTallyFor(ev.Actor)
				return tally.Total-tally.Noncreature == 2
			}, label, effects.ReturnThisCardFromYourGraveyard), label+"?")),
		},
	})
}
