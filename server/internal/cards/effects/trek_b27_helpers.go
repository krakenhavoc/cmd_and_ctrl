package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b27_helpers.go — Star Trek, batch 27 (#4973): the pieces the
// batch's cards share.

// trekB27TuckTargetCardToBottom is "Put target card from your graveyard on
// the bottom of your library" (Cybernetic Specialist): the still-legal
// target card goes to the bottom of its owner's library (CR 608.2b).
func trekB27TuckTargetCardToBottom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return g.TuckToLibraryForEffect(t.ID, true)
		}
	}
	return nil
}

// trekB27CreaturesYouControlGet is a "Creatures you control get +P/+T."
// static switched on by a station threshold (Deep Space 9's 9+ line,
// CR 721.2a): b16Anthem over the source controller's creatures, gated on
// the charge-counter count and read live.
func trekB27CreaturesYouControlGet(label string, n, power, toughness int) game.StaticAbility {
	s := b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller
	}, power, toughness)
	s.Label = label
	s.ActiveWhen = AtChargeCounters(n)
	return s
}

// trekB27ExileTopFourPutPermanents is "exile the top four cards of your
// library. Put all permanent cards exiled this way onto the battlefield"
// (Deep Space 9). The exile is MillToZone to exile, so the cards that put
// the tail are the ones that actually reached exile (CR 400.7); the tail
// puts every permanent card among them that is still there onto the
// battlefield under the ability's controller in one simultaneous entry
// (CR 110.2a). Nonpermanent cards stay in exile.
func trekB27ExileTopFourPutPermanents(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return MillToZone{Player: item.Controller, N: 4, To: game.ZoneExile, ThenTail: ctx.ItemTail(trekB27PutExiledPermanents)}.Apply(ctx)
}

// trekB27PutExiledPermanents is the put half. An on-disk key.
var trekB27PutExiledPermanents = CardTail("trek-b27/deep-space-9-put-permanents", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	var entries []game.BatchEntry
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if z := g.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneExile || !isPermanentCard(c) {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneExile})
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: ctx.Controller()}, game.KeyedTail{})
})

// trekB27OnePlusLifeGainedThisTurn is Sickbay's X: one plus the life the
// trigger's controller gained this turn, read as the trigger resolves.
func trekB27OnePlusLifeGainedThisTurn(g *game.Game, controller uuid.UUID) int {
	return 1 + b15LifeGainedThisTurn(g, controller)
}

// modeSourceDealsDamageToTargetPlayer is a modal bullet's "<this> deals N
// damage to target player/opponent": the bullet's target player, if it is
// still legal, is dealt N damage by the ability's source (its last-known
// information if it has left, CR 608.2h). Hawkeye, Master Marksman's
// Explosive and Tea, Earl Grey, Hot's first mode.
func modeSourceDealsDamageToTargetPlayer(n int) func(item *game.StackItem, ctx *Context, occ int) error {
	return func(item *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
}
