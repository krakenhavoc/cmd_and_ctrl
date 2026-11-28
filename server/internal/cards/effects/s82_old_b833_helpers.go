package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ice Age activated abilities, batch 833 (#5636): helpers the cards of the
// batch read.

// OldB833BasicLand is "basic land" as a target predicate, read off the effective
// supertypes.
func OldB833BasicLand(extra ...CardPredicate) CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		if !IsBasicLand(c) {
			return false
		}
		for _, p := range extra {
			if !p(g, caster, c) {
				return false
			}
		}
		return true
	}
}

// OldB833TargetSnowChange is Arcum's Weathervane's body: the target land is
// no longer snow (remove), or becomes snow (add), with no stated duration, so
// it lasts indefinitely (CR 611.2a). A land that has left is skipped.
func OldB833TargetSnowChange(name string, add bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := FirstLegalBattlefieldTarget(ctx)
		if id == uuid.Nil {
			return nil
		}
		mod, label := game.RemoveSupertypesMod("Snow"), name+" — no longer snow"
		if add {
			mod, label = game.AddSupertypesMod("Snow"), name+" — becomes snow"
		}
		return ScopedEffectFor{Target: id, Mods: []game.Mod{mod}, Duration: game.IndefiniteDuration(), Label: label}.Apply(ctx)
	}
}

// OldB833PreventNextToTheTarget is Elvish Healer's "Prevent the next 1 damage
// that would be dealt to any target this turn. If it's a green creature,
// prevent the next 2 damage instead." The target is read as it resolves.
func OldB833PreventNextToTheTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ts := ctx.LegalTargets()
		if len(ts) == 0 {
			return nil
		}
		n := 1
		if ts[0].Kind == game.TargetCard {
			if c, ok := g.LookupCardForEffect(ts[0].ID); ok && c.IsCreature() && c.HasColor(ctx.TextColor("G")) {
				n = 2
			}
		}
		label := name + ": prevent the next " + strconv.Itoa(n) + " damage"
		return PreventNextDamage{Target: ts[0].ID, Amount: n, Label: label}.Apply(ctx)
	}
}

// oldB833ReturnTheObjectToHandBody is the delayed "return that creature to
// its owner's hand" of Barbarian Guides: the object named, if it is still
// that object on the battlefield (CR 603.7c). An on-disk key.
var oldB833ReturnTheObjectToHandBody = game.DelayedBody("bounce/old-b833-the-object", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !OnBattlefield(g, p.Object.ID) {
		return nil
	}
	return BounceToHand{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// OldB833ChooseLandTypeThenSnowwalk asks for a land type (CR 205.3i), then
// the target creature gains snow landwalk of that type until end of turn and
// is returned to its owner's hand at the beginning of the next end step.
func OldB833ChooseLandTypeThenSnowwalk(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		then := ctx.ItemTail(oldB833SnowwalkChosen)
		then.Words = []string{name}
		return ChooseBasicLandTypeTail(ctx, name+" — choose a land type", game.LandTypes, then)
	}
}

var oldB833SnowwalkChosen = CardTail("then/old-b833-snowwalk-type-chosen", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Words) == 0 || o.Tail.Word == "" {
		return nil
	}
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	name := o.Tail.Words[0]
	token := game.SnowLandwalkToken(o.Tail.Word)
	if token == "" {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{token}, Label: name + " — " + token}).Apply(ctx); err != nil {
		return err
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  name + " — return " + c.Name + " to its owner's hand",
		Cards:  []uuid.UUID{id},
		Body:   oldB833ReturnTheObjectToHandBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
