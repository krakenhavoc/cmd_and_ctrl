package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b299_helpers.go — shared pieces of the Oath of the Gatewatch
// batch 299 cards (#5545).

// OldB299ThreeCountersAndAnimateTheLand is "you may put three +1/+1 counters
// on target land you control. If you do, that land becomes a 0/0 Elemental
// creature with haste that's still a land" (Cyclone Sire, Wall of Resurgence)
// as the body of an optional, targeted trigger: awaken's verb (CR 702.113a)
// over the still-legal target land, with no duration (CR 611.2a).
func OldB299ThreeCountersAndAnimateTheLand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return g.AwakenForEffect(ctx.Controller(), ctx.Source(), t.ID, 3)
		}
	}
	return nil
}

// OldB299HedronAlignmentReveal is Hedron Alignment's upkeep trigger as it
// resolves: reveal your hand, then win the game if you own a card named Hedron
// Alignment in each of exile, your hand, your graveyard and on the battlefield.
func OldB299HedronAlignmentReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := ctx.Controller()
	g.RevealHandForEffect(me)
	const name = "Hedron Alignment"
	owned := func(c game.Card) bool { return c.Owner == me && c.Name == name }
	var inExile, inHand, inGraveyard, onBattlefield bool
	for _, c := range g.Exile.Cards {
		inExile = inExile || owned(c)
	}
	for _, id := range AllHandCardIDs(g, me) {
		if c, ok := g.LookupCardForEffect(id); ok {
			inHand = inHand || owned(c)
		}
	}
	if p := g.PlayerByIDForEffect(me); p != nil {
		for _, c := range p.Graveyard.Cards {
			inGraveyard = inGraveyard || owned(c)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		onBattlefield = onBattlefield || owned(c)
	}
	if !(inExile && inHand && inGraveyard && onBattlefield) {
		return nil
	}
	return WinTheGame{Player: me}.Apply(ctx)
}
