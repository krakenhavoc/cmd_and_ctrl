package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Colfenor's Urn — Artifact {3}:
//
//	"Whenever a creature with toughness 4 or greater is put into your
//	 graveyard from the battlefield, you may exile it.
//	 At the beginning of the end step, if three or more cards have been
//	 exiled with this artifact, sacrifice it. If you do, return those cards
//	 to the battlefield under their owner's control."
//
// The two abilities are linked (CR 607.2a). The first reads the creature's
// toughness as it last was on the battlefield (CR 603.10a) and only fires for
// a card going to YOUR graveyard (a creature you own), then exiles it linked
// to this Urn object (ExileTarget.Linked) if it is still in the graveyard.
// The second is "each end step" with an intervening "if" (CR 603.4): it
// counts the cards linked to this object when the step begins and again as
// it resolves, sacrifices the Urn, and only if the sacrifice happened returns
// the cards that were exiled with it, each under its owner's control. A
// card that has left exile since is a new object and is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "920c4df3-9517-44cd-8304-28b1ef69d60e",
		Name:         "Colfenor's Urn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.DiedCreature(ev, g)
				return ok && c.Owner == source.Controller && effects.B29LastKnownToughnessOffBattlefield(g, ev.CardID) >= 4
			}, "Colfenor's Urn — you may exile the creature", colfenorsUrnExile),
				"Colfenor's Urn — exile that creature?"),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15EndStepBegan(ev) && len(g.ExiledWithForEffect(game.ObjectRefOf(*source))) >= 3
			}, "Colfenor's Urn — if three or more cards have been exiled with it, sacrifice it and return those cards to the battlefield under their owners' control",
				colfenorsUrnReturn),
		},
	})
}

// colfenorsUrnExile exiles the creature that died, linked to this Urn, if it
// is still in the graveyard it was put into.
func colfenorsUrnExile(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ExileTarget{Target: id, Linked: true}.Apply(effects.NewContext(g, item))
}

// colfenorsUrnReturn is the end-step ability: recheck the count (CR 603.4),
// then sacrifice the Urn; the cards ride the sacrifice's tail.
func colfenorsUrnReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) < 3 {
		return nil
	}
	return effects.SacrificePermanent{Target: item.SourceCardID, ThenTail: effects.WithCarry(ctx.ItemTail(colfenorsUrnReturned), ids...)}.Apply(ctx)
}

// colfenorsUrnReturned is "if you do, return those cards": the Urn was
// sacrificed, so every card exiled with it that is still in exile enters
// under its owner's control. An on-disk key: never renamed, never reused.
var colfenorsUrnReturned = effects.CardTail("colfenors-urn/return", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	for _, id := range o.Tail.Carry {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if _, err := ctx.Game.ReturnFromExileToBattlefieldForEffect(id, c.Owner, false); err != nil {
			return err
		}
	}
	return nil
})
