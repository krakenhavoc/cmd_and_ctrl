package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB238EternalizeDiscard is Eternalize whose cost also says "Discard a card"
// (Hour of Devastation's Sunscourge Champion and Sinuous Striker): the mana,
// a card from your hand, and exiling this card from your graveyard are one
// cost (CR 702.129a), the discard picked at announce (CR 602.2b). The token is
// Eternalize's own.
func OldB238EternalizeDiscard(cost string) ActivatedAbility {
	ab := Eternalize(cost)
	ab.Label = "Eternalize—" + cost + ", Discard a card (" + cost + ", Discard a card, Exile this card from your graveyard: " +
		"Create a token that's a copy of it, except it's a 4/4 black Zombie in addition to its other types and it has no mana cost. " +
		"Activate only as a sorcery.)"
	ab.Cost = Plus(ManaCost(cost), DiscardACard(), ExileThis())
	return ab
}

// OldB238CardTargetLastKnown is the first card target of the item still legal
// as it resolves (CR 608.2b) and the card as it last existed (CR 608.2h), read
// before an effect moves it.
func OldB238CardTargetLastKnown(ctx *Context) (game.Card, bool) {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ctx.Game.LookupCardForEffect(t.ID)
	}
	return game.Card{}, false
}

// OldB238DefendingGraveyardCardTarget is the clause "target card from defending
// player's graveyard" of an attack trigger, read off the source's attack:
// the player it is attacking (Hive of the Eye Tyrant, Graven Abomination).
func OldB238DefendingGraveyardCardTarget(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "target card from defending player's graveyard",
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.Owner == defender
		},
		Min: 1, Max: 1,
	}
}

// OldB238ExileEachLegalTargetCard is "exile <the targets>" for an ability whose
// targets are cards: each one still legal as it resolves (CR 608.2b).
func OldB238ExileEachLegalTargetCard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range LegalTargetCards(item, g) {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
