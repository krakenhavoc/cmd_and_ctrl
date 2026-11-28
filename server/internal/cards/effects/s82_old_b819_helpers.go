package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b819_helpers.go — Alliances, damage, token makers, enters and dies
// triggers (batch 819).

// OldB819CreatureCardDirectlyAbove is "this card is in your graveyard with a
// creature card directly above it" (Death Spark, Krovikan Horror): the card
// sits in its owner's graveyard and the card put into that graveyard right
// after it, the one on top of it in the pile, is a creature card.
func OldB819CreatureCardDirectlyAbove(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return false
	}
	p := g.PlayerByIDForEffect(c.Owner)
	if p == nil || p.Graveyard == nil {
		return false
	}
	cards := p.Graveyard.Cards
	for i := range cards {
		if cards[i].InstanceID != id {
			continue
		}
		return i+1 < len(cards) && cards[i+1].IsCreature()
	}
	return false
}

// OldB819GraveyardCreatureAboveCondition is the intervening "if" of those two
// triggers (CR 603.4): checked as the ability triggers and again as it
// resolves.
var OldB819GraveyardCreatureAboveCondition = CondBoard{
	Name: "this card is in your graveyard with a creature card directly above it",
	Test: func(ctx *Context) bool { return OldB819CreatureCardDirectlyAbove(ctx.Game, ctx.Source()) },
}

// OldB819InGraveyardUnderACreature is that same condition as an event
// predicate for the trigger's own filter.
func OldB819InGraveyardUnderACreature(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return OldB819CreatureCardDirectlyAbove(g, source.InstanceID)
}

// OldB819ReturnThisFromGraveyardToHand moves the source card from its owner's
// graveyard to its hand, if it is still there.
func OldB819ReturnThisFromGraveyardToHand(ctx *Context, source uuid.UUID) error {
	z := ctx.Game.FindCardZoneForEffect(source)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: source, Dest: game.ZoneHand}.Apply(ctx)
}

// OldB819WasInCombat is "this creature attacked or blocked this combat":
// read from this turn's attackers and the combat's blocks.
func OldB819WasInCombat(g *game.Game, ref game.ObjectRef) bool {
	if slices.Contains(g.AttackersThisCombatForEffect(), ref) {
		return true
	}
	for _, b := range g.TurnTally.CombatBlocks {
		if b.Blocker == ref && b.PhaseID == g.Turn.PhaseID {
			return true
		}
	}
	return false
}

// OldB819TimesRegeneratedThisTurn counts this turn's regenerations of `id`
// (CR 701.19): one event per regeneration shield used.
func OldB819TimesRegeneratedThisTurn(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventRegenerated && ev.CardID == id {
			n++
		}
	}
	return n
}
