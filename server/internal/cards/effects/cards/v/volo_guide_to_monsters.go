package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volo, Guide to Monsters — Legendary Creature — Human Wizard {2}{G}{U}, 3/2
// (Adventures in the Forgotten Realms):
//
//	"Whenever you cast a creature spell that doesn't share a creature type
//	 with a creature you control or a creature card in your graveyard, copy
//	 that spell. (A copy of a creature spell becomes a token.)"
//
// The "doesn't share" clause is part of what the trigger watches for, not an
// intervening "if": it is judged as the spell is cast (the spell is on the
// stack, so it is not among "a creature you control"), and is not asked again
// as the trigger resolves. Changelings share a type with every creature that
// has one (CR 702.73a), which game.SharesCreatureType reads. The copy is made
// from last-known information if the spell has left the stack by then
// (CR 608.2h) and becomes a token as it resolves (CR 608.3f); the copy is not
// "cast", so Volo does not trigger on it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f105960e-6609-4192-afc0-e243013b7cc1",
		Name:         "Volo, Guide to Monsters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, voloSharesNoType,
				"Volo, Guide to Monsters — copy that creature spell",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CopySpell{
						StackID:       ctx.Trigger().Event.CardID,
						Controller:    ctx.Controller(),
						FromLastKnown: true,
					}.Apply(ctx)
				}),
		},
	})
}

// voloSharesNoType: you cast a creature spell, and it shares no creature type
// with a creature you control or a creature card in your graveyard.
func voloSharesNoType(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !spell.IsCreature() {
		return false
	}
	for _, p := range g.BattlefieldCardsForEffect() {
		if p.Controller == source.Controller && p.IsCreature() && game.SharesCreatureType(&spell, &p) {
			return false
		}
	}
	if me := g.PlayerByIDForEffect(source.Controller); me != nil && me.Graveyard != nil {
		for _, c := range me.Graveyard.Cards {
			if c.IsCreature() && game.SharesCreatureType(&spell, &c) {
				return false
			}
		}
	}
	return true
}
