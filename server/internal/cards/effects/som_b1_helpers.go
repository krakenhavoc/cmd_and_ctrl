package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// som_b1_helpers.go — the shapes Scars of Mirrodin, batch 1 (#5577)
// prints more than once. Append-only.

// somB1Metalcraft is metalcraft's condition, "as long as you control three
// or more artifacts" (CR 207.2c names it an ability word, so it has no
// rules meaning of its own): the source's controller's artifacts, the
// source itself included when it is one. It runs inside a layer
// recompute, so it reads the battlefield in place.
func somB1Metalcraft(g *game.Game, source *game.Card) bool {
	return CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() }) >= 3
}

// somB1MetalcraftCost is the same condition for a cost reduction (Stoic
// Rebuttal), counted as the spell is cast.
func SomB1MetalcraftCost() CostPredicate {
	return func(q game.CostQuery) bool {
		return PermanentsYouControl(Artifact())(q) >= 3
	}
}

// somB1MetalcraftPump is "Metalcraft — This creature gets +P/+T as long as
// you control three or more artifacts." (Chrome Steed): Drover of the
// Mighty's layer 7c self pump on the metalcraft gate. It is not locked
// in (CR 611.3a): it turns on and off as the count crosses three.
func SomB1MetalcraftPump(label string, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return SelfOnly(target, g, source) && somB1Metalcraft(g, source)
		},
		Apply: PumpBy(power, toughness),
	}
}

// somB1SelfGainsUntilEOT is "{cost}: This creature gains <keyword> until
// end of turn." (Wall of Tanglecord, Vector Asp). An activation whose
// creature has left, or come back as a new object (CR 400.7), does
// nothing.
func SomB1SelfGainsUntilEOT(name, cost, keyword string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": This creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
		Cost:    ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !SourceIsStillThisPermanent(g, item) {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target: item.SourceCardID, Keywords: []string{keyword},
				Label: name + " — " + keyword + " until end of turn",
			}.Apply(NewContext(g, item))
		},
	}
}

// somB1CounterTheTargetSpells is "Counter target spell" as a spell's
// resolution (Turn Aside, Stoic Rebuttal, Rebuff the Wicked): each target
// spell still legal is countered (CR 608.2b).
func SomB1CounterTheTargetSpells(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// somB1UntapEachOther is "untap each other <permanent> you control"
// (Copperhorn Scout's creatures, Myr Galvanizer's Myr). The set is read
// once as the ability resolves; the source is never in it, whether or not
// it is still on the battlefield, and each untap is a member of one group
// (CR 608.2h).
func SomB1UntapEachOther(match func(ctx *Context, c game.Card) bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID == item.SourceCardID || c.Controller != item.Controller || !c.Tapped || !match(ctx, c) {
				continue
			}
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if err := (UntapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
