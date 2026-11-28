package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ice Age pump and combat tricks, batch 834 (#5636): helpers the cards of the
// batch read.

// OldB834PowerToughness is "<P>/<T> creature" as a target predicate: the
// creature's current power and toughness are exactly these (Aegis of the Meek).
func OldB834PowerToughness(power, toughness int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.CurrentPower() == power && c.CurrentToughness() == toughness
	}
}

// OldB834TargetGets is "Target creature gets +P/+T until end of turn" as an
// activated ability's body. A target that has left gets nothing (CR 608.2b).
func OldB834TargetGets(power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := FirstLegalBattlefieldTarget(ctx)
		if id == uuid.Nil {
			return nil
		}
		return BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
	}
}

// OldB834BattleFrenzy is "Green creatures you control get +1/+1 until end of
// turn. Nongreen creatures you control get +1/+0 until end of turn." The two
// sets are fixed as it resolves (CR 611.2c).
func OldB834BattleFrenzy(_ *game.StackItem, ctx *Context) error {
	mine := And(Creature(), YouControl())
	if err := (BoostUntilEOT{Match: And(mine, OfColor("G")), Power: 1, Toughness: 1, Label: "Battle Frenzy — green creatures you control get +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	return BoostUntilEOT{Match: And(mine, Not(OfColor("G"))), Power: 1, Label: "Battle Frenzy — nongreen creatures you control get +1/+0"}.Apply(ctx)
}

// OldB834Stampede is "Attacking creatures get +1/+0 and gain trample until end
// of turn": every attacking creature, whoever controls it, fixed as it
// resolves.
func OldB834Stampede(_ *game.StackItem, ctx *Context) error {
	if err := (BoostUntilEOT{Match: AttackingCreature(), Power: 1, Label: "Stampede — attacking creatures get +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Match: AttackingCreature(), Keywords: []string{"trample"}, Label: "Stampede — trample"}.Apply(ctx)
}

// OldB834OtherCombatantsGetPerOther is Márton Stromgald's pump: each other
// creature matching `side` (attacking, or blocking) gets +N/+N until end of
// turn, N being the number of such creatures other than the source. The set is
// fixed as the trigger resolves (CR 611.2c).
func OldB834OtherCombatantsGetPerOther(side CardPredicate, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		others := And(side, OtherThan(item.SourceCardID))
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if others(g, item.Controller, c) {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return BoostUntilEOT{Match: others, Power: n, Toughness: n, Label: label}.Apply(ctx)
	}
}

// oldB834SacrificeTheObjectBody is the delayed "sacrifice it" of a creature
// the ability named: the object, if it is still that object on the
// battlefield under the delayed trigger's controller (CR 603.7c, 701.21a).
// An on-disk key.
var oldB834SacrificeTheObjectBody = game.DelayedBody("sacrifice/old-b834-the-object", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !OnBattlefield(g, p.Object.ID) || c.Controller != item.Controller {
		return nil
	}
	return SacrificePermanent{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// OldB834TargetFliesThenSacrifice is Krovikan Elementalist's second ability:
// "Target creature you control gains flying until end of turn. Sacrifice it at
// the beginning of the next end step."
func OldB834TargetFliesThenSacrifice(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := FirstLegalBattlefieldTarget(ctx)
		if id == uuid.Nil {
			return nil
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: label + " — flying"}).Apply(ctx); err != nil {
			return err
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  label + " — sacrifice " + c.Name,
			Cards:  []uuid.UUID{id},
			Body:   oldB834SacrificeTheObjectBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}.Apply(ctx)
	}
}

// OldB834PhantasmalMount is Phantasmal Mount's ability: the target creature
// you control (toughness 2 or less) gets +1/+1 and gains flying until end of
// turn, and for this turn each of the pair leaving the battlefield has the
// other sacrificed: two delayed triggers, one naming the Mount and one the
// creature, each pinned to the object as it is now (CR 603.7c, 400.7).
func OldB834PhantasmalMount(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: "Phantasmal Mount — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: "Phantasmal Mount — flying"}).Apply(ctx); err != nil {
		return err
	}
	tc, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	targetRef := game.ObjectRef{ID: target, Epoch: tc.ObjectEpoch}
	d := g.PinnedTo(g.UntilEndOfTurnDuration(), target)
	// "When this creature leaves the battlefield this turn, sacrifice that
	// creature."
	mount := item.SourceCardID
	if mc, ok := g.LookupCardForEffect(mount); ok && mount != target {
		mountDur := g.PinnedTo(g.UntilEndOfTurnDuration(), mount)
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   ctx.Controller(),
			SourceCardID: mount,
			Label:        "Phantasmal Mount — sacrifice that creature",
			On:           []game.EventKind{game.EventLTB},
			Condition:    theListedObjectLeftCondition,
			Cards:        []uuid.UUID{mount},
			Duration:     &mountDur,
			Body:         oldB834SacrificeTheObjectBody,
			Params:       game.EffectParams{Object: targetRef},
		})
		// "When the creature leaves the battlefield this turn, sacrifice this
		// creature."
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   ctx.Controller(),
			SourceCardID: mount,
			Label:        "Phantasmal Mount — sacrifice this creature",
			On:           []game.EventKind{game.EventLTB},
			Condition:    theListedObjectLeftCondition,
			Cards:        []uuid.UUID{target},
			Duration:     &d,
			Body:         oldB834SacrificeTheObjectBody,
			Params:       game.EffectParams{Object: game.ObjectRef{ID: mount, Epoch: mc.ObjectEpoch}},
		})
	}
	return nil
}

// OldB834ChaosMoon is Chaos Moon's upkeep trigger: count the permanents. Odd:
// until end of turn red creatures get +1/+1 and whenever a player taps a
// Mountain for mana that player adds an additional {R}. Even: red creatures
// get -1/-1 and a Mountain tapped for mana produces colorless mana instead of
// any other type. The red creatures are fixed as it resolves (CR 611.2c).
func OldB834ChaosMoon(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	red := And(Creature(), OfColor("R"))
	if len(g.BattlefieldCardsForEffect())%2 == 1 {
		if err := (BoostUntilEOT{Match: red, Power: 1, Toughness: 1, Label: "Chaos Moon — red creatures get +1/+1"}).Apply(ctx); err != nil {
			return err
		}
		return TurnManaTrigger{Label: "Chaos Moon — adds an additional {R}", Subtype: "Mountain", Adds: "{R}"}.Apply(ctx)
	}
	if err := (BoostUntilEOT{Match: red, Power: -1, Toughness: -1, Label: "Chaos Moon — red creatures get -1/-1"}).Apply(ctx); err != nil {
		return err
	}
	g.LandSubtypeProducesColorlessThisTurnForEffect(item.SourceCardID, ctx.Controller(), ctx.TextSubtype("Mountain"),
		"Chaos Moon — Mountains tapped for mana produce colorless mana")
	return nil
}
