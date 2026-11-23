package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82SncB11Treasures is how many Treasures `player` controls.
func S82SncB11Treasures(g *game.Game, player uuid.UUID) int {
	return len(PermanentsControlledByMatching(g, player, Subtype("Treasure")))
}

// s82SncB11PumpOrCounter is the body of Jetmir's Fixer and Paragon of
// Modernity: the source gets +1/+1 until end of turn, or, when counter
// is true, a +1/+1 counter instead. Nothing when the source has left
// or is a new object (CR 400.7).
func S82SncB11PumpOrCounter(label string, counter func(ctx *Context) bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if counter(ctx) {
			return ThisGetsAPlusOneCounter(g, item)
		}
		return ThisGetsUntilEndOfTurn(1, 1, label)(g, item)
	}
}

// OncePerBatchPerAttacker is OncePerBatch for a clause that names the
// attacking PLAYER ("whenever a player attacks one or more of your
// opponents"): the batch collapses per attacker, so one declaration is
// one trigger however many creatures and defenders it has.
func OncePerBatchPerAttacker(t game.TriggeredAbility) game.TriggeredAbility {
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Actor.String() }
	return t
}

// s82SncB11ControlsACreatureThatFought is "you control a creature that
// fought this turn" (Boxing Ring): a creature you control now has an
// EventFight this turn that its CURRENT object made, so a creature that
// left the battlefield and came back after fighting is a new object and
// does not count (CR 400.7). EventFight carries the fighter as CardID
// and no epoch, so an entry after the fight is read as a new object.
func S82SncB11ControlsACreatureThatFought(g *game.Game, controller, _ uuid.UUID) bool {
	events := g.EventsThisTurn()
	for i, ev := range events {
		if ev.Kind != game.EventFight {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.Controller != controller || !c.IsCreature() || !g.Battlefield.Contains(ev.CardID) {
			continue
		}
		reentered := false
		for _, later := range events[i+1:] {
			if later.Kind == game.EventETB && later.CardID == ev.CardID {
				reentered = true
				break
			}
		}
		if !reentered {
			return true
		}
	}
	return false
}

// s82SncB11EnteredFights is Boxing Ring's body: the creature that
// entered fights the first target that is still legal.
func S82SncB11EnteredFights(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		other, ok := g.PermanentRefForEffect(t.ID)
		if !ok {
			return nil
		}
		return g.FightForEffect(obj.Ref(), other)
	}
	return nil
}

// s82SncB11VaziCounterThenDraw is "put a +1/+1 counter on target creature,
// then draw a card": nothing at all when the target is no longer legal.
func S82SncB11VaziCounterThenDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return DrawCards{N: 1}.Apply(ctx)
	}
	return nil
}
