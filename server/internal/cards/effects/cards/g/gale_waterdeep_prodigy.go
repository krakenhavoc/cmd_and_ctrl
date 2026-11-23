package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gale, Waterdeep Prodigy — Legendary Creature — Human Wizard {2}{U}, 1/3
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Whenever you cast an instant or sorcery spell from your hand, you may
//	 cast up to one target card of the other type from your graveyard. If a
//	 spell cast from your graveyard this way would be put into your
//	 graveyard, exile it instead.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// The trigger fires as the spell becomes cast, from the hand (the event's
// OldZone). "The other type" is read off that spell as the ability is put on
// the stack: an instant spell offers sorcery cards, a sorcery spell offers
// instant cards. The cast itself is Emet-Selch's OfferGraveyardCast: made as
// the ability resolves (CR 608.2g) at the card's mana cost, the spell exiled
// instead of going to the graveyard whether it resolves, is countered or
// fizzles; a target that left the graveyard is not cast.
//
// Choose a Background is a deck-construction rule (CR 702.124k) that
// internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	const label = "Gale, Waterdeep Prodigy — you may cast up to one target card of the other type from your graveyard"
	trig := effects.On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ev.OldZone == game.ZoneHand &&
			effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery()))(ev, source, lki, g)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return effects.OfferGraveyardCast{Target: t.ID, SourceName: "Gale, Waterdeep Prodigy"}.Apply(ctx)
		}
		return nil
	})
	trig.TargetsFrom = func(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		spell, ok := g.LookupCardForEffect(tc.Event.CardID)
		if !ok {
			return nil
		}
		other := effects.Sorcery()
		if effects.Sorcery()(g, source.Controller, spell) {
			other = effects.Instant()
		}
		return effects.TargetCardInGraveyard("up to one target card of the other type from your graveyard",
			other, effects.YouOwn()).WithCount(0, 1)
	}
	effects.Register(effects.Spec{
		OracleID:     "cca15007-2faf-4696-a4c5-2d7b6c1ec5b5",
		Name:         "Gale, Waterdeep Prodigy",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
