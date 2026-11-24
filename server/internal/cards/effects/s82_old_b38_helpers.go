package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b38_helpers.go — shared bodies for the S82 older-sets batch 38
// (Modern Horizons 2: delirium, storm, Auras; #5501).

// S82OldB38CascadeWithDelirium is "Delirium — This spell has cascade as
// long as there are four or more card types among cards in your
// graveyard" (Bloodbraid Marauder): the cascade keyword's own trigger,
// gated on delirium as the spell is cast (CR 702.85a), the graveyard
// read being the caster's.
func S82OldB38CascadeWithDelirium() game.TriggeredAbility {
	t := Cascade()
	base := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
		return base(ev, source, c, g) && B16CardTypesInGraveyard(g, ev.Actor) >= 4
	}
	return t
}

// S82OldB38ControlsAToken is the condition "you control a token".
func S82OldB38ControlsAToken() Condition {
	return CondBoard{Name: "you control a token", Test: func(ctx *Context) bool {
		return ControlsAnyCard(ctx.Game, ctx.Controller(), IsTokenPredicate())
	}}
}

// S82OldB38DamageEachCreatureDefendingPlayerControls is Gouged Zealot's
// body: 1 damage from the attacker to each creature the defending player
// controls, the set fixed as the ability resolves (CR 608.2).
func S82OldB38DamageEachCreatureDefendingPlayerControls(amount int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		defender := DefendingPlayerOfAttack(g, item.Trigger.Event)
		if defender == uuid.Nil {
			return nil
		}
		return B11DamageEachCreatureControlledBy(g, item, defender, amount)
	}
}

// S82OldB38ReturnThisToOwnersHand is "Return <this> to its owner's hand"
// for a card that may be in either zone ("activate only if it is on the
// battlefield or in your graveyard", Skyblade's Boon).
func S82OldB38ReturnThisToOwnersHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneBattlefield:
		return BounceToHand{Target: item.SourceCardID}.Apply(ctx)
	case game.ZoneGraveyard:
		return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(ctx)
	}
	return nil
}

// S82OldB38GrantFlyingToTargets gives every still-legal target flying
// until end of turn (Ghost-Lit Drifter's two abilities).
func S82OldB38GrantFlyingToTargets(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
