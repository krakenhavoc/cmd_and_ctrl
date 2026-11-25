package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b165_helpers.go — shared vocabulary of Modern Horizons batch 165
// (S82 older sets, #5519).

// OldB165HellbentTribalGrant is "Hellbent — As long as you have no cards in
// hand, <tribe> creatures you control have <ability>": the bundle grant to the
// tribe's creatures its controller has, held to an empty hand and re-read
// whenever a hand changes size.
func OldB165HellbentTribalGrant(f TribeFilter, key string) game.StaticAbility {
	ab := GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return f.Matches(target, g, source) && R35YouHaveNoCardsInHand(g, source.Controller, uuid.Nil)
	}, key)
	ab.Label = "Hellbent — " + f.subject() + " have additional abilities as long as you have no cards in hand."
	ab.DependsOnHandSize = true
	return ab
}

// OldB165OnlyCreatureCardInYourGraveyard is Nether Spirit's intervening "if"
// (CR 603.4): the card the ability belongs to is the only creature card in its
// owner's graveyard. It holds as the ability triggers and again as it
// resolves, and a card that has left the graveyard is no longer there to be
// the only one.
func OldB165OnlyCreatureCardInYourGraveyard(g *game.Game, owner, card uuid.UUID) bool {
	p := g.PlayerByIDForEffect(owner)
	if p == nil || p.Graveyard == nil {
		return false
	}
	found := false
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() {
			continue
		}
		if c.InstanceID != card || found {
			return false
		}
		found = true
	}
	return found
}

// OldB165IfThisIsTheOnlyCreatureCardInYourGraveyard narrows a graveyard
// trigger to Nether Spirit's intervening "if".
func OldB165IfThisIsTheOnlyCreatureCardInYourGraveyard(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && OldB165OnlyCreatureCardInYourGraveyard(g, source.Owner, source.InstanceID) &&
			(when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "this card is the only creature card in your graveyard", Test: func(ctx *Context) bool {
		return OldB165OnlyCreatureCardInYourGraveyard(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
	}}, t)
}
