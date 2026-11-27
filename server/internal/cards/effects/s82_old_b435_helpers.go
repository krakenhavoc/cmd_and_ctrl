package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// UntappedCreatureYouControl is "untapped creatures you control" as a static's
// subject (Builder's Blessing): a creature the source's controller controls
// that is untapped as the layers are applied.
func UntappedCreatureYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller && !target.Tapped
}

// CreatureYouControlTargetedByAnInstantOrSorcerySpell is "whenever a creature
// you control becomes the target of an instant or sorcery spell" (Wild
// Defiance). The targeter must be a spell on the stack, and an instant or
// sorcery one; an ability never counts.
func CreatureYouControlTargetedByAnInstantOrSorcerySpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !B40CreatureYouControlTargetedByASpell(ev, source, g) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.Source)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// PumpTargetAndItsSoulbondPartner is Joint Assault's body: the target gets
// +2/+2 until end of turn and, if it is paired with a creature, that creature
// also gets +2/+2 until end of turn. The partner is read as the spell
// resolves.
func PumpTargetAndItsSoulbondPartner(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	partner := ctx.Game.PairedPartner(&c)
	if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Joint Assault — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	if partner == nil {
		return nil
	}
	return BoostUntilEOT{Target: partner.InstanceID, Power: 2, Toughness: 2, Label: "Joint Assault — paired creature +2/+2"}.Apply(ctx)
}

// ThatcherRevoltTokens is Thatcher Revolt's body: three 1/1 red Human
// creature tokens with haste, sacrificed at the beginning of the next end step.
func ThatcherRevoltTokens(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids, err := g.CreateTokensForEffect(item.Controller, ctx.TextToken(TokenCard("1/1 red Human")), 3,
		game.TokenEntryOptions{Keywords: []string{"haste"}})
	if err != nil || len(ids) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: item.Controller,
		Label:      "Thatcher Revolt — sacrifice the tokens",
		Cards:      ids,
		Body:       SacrificeListedCardsBody,
	}.Apply(ctx)
}

// LairDelveBody is Lair Delve: reveal the top two cards of your library, put
// all creature and land cards revealed this way into your hand and the rest on
// the bottom of your library in any order.
func LairDelveBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := item.Controller
	revealed := g.RevealTopOfLibraryForEffect(player, item.SourceCardID, 2, "Lair Delve — reveal the top two cards of your library")
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		Match:  Or(Creature(), Land()),
		All:    true,
		Label:  "Lair Delve — put all creature and land cards into your hand, the rest on the bottom",
		Key:    TakeRestBottomAny,
	}.Apply(ctx)
}
