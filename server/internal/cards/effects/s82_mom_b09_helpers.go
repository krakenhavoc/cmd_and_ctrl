package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers shared by the Knight cards of S82's March of the Machine batch 9.

// s82Mom09KnightEnteredUnderYourControl is "another Knight you control
// enters": an EventETB for a permanent other than the source, now a Knight
// and controlled by the source's controller (CR 603.6a).
func S82Mom09AnotherKnightEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	return B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Knight"))
}

// s82Mom09OtherKnightsControlled counts the Knights source's controller
// controls other than the source itself, read through the layered view.
func s82Mom09OtherKnightsControlled(g *game.Game, source game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller &&
			c.IsCreature() && c.HasSubtype(source.TextSubtype("Knight")) {
			n++
		}
	}
	return n
}

// s82Mom09AnotherKnightEnteredThisTurn is Swordsworn Cavalier's condition:
// another Knight entered the battlefield under the controller's control this
// turn. The per-turn tally counts every Knight that entered, so the source's
// own entry (if it entered this turn) is taken off before asking.
func S82Mom09AnotherKnightEnteredThisTurn(g *game.Game, source *game.Card) bool {
	n := g.EnteredWithSubtypeThisTurn(source.Controller, source.TextSubtype("Knight"))
	n -= g.TurnTally.Entered[source.InstanceID]
	return n > 0
}

// s82Mom09TapAndMaybeStun is Protocol Knight's body: tap the target, then
// put a stun counter on it if the controller controls another Knight, checked
// as the ability resolves (it is not an intervening "if").
func S82Mom09TapAndMaybeStun(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	id := ts[0].ID
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		// The Knight left before this resolved: last known information.
		src = game.Card{InstanceID: item.SourceCardID, Controller: item.Controller, TypeLine: "Creature — Human Knight"}
	}
	if s82Mom09OtherKnightsControlled(g, src) == 0 {
		return nil
	}
	return (AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx)
}

// s82Mom09KnightDealtCombatDamage is "a Knight you control deals combat
// damage to a player".
func S82Mom09KnightDealtCombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !CombatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && OfCreatureType("Knight")(g, source.Controller, game.ThroughTextOf(source, c))
}

// s82Mom09CounterOnThatCreature puts a +1/+1 counter on the creature that
// dealt the combat damage; one that has left does nothing (CR 608.2b).
func S82Mom09CounterOnThatCreature(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.Source
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(NewContext(g, item))
}

// s82Mom09NontokenKnightYouControlDied is "a nontoken Knight you control
// dies", read from the creature's last-known information (CR 603.10a).
func S82Mom09NontokenKnightYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && !IsToken(dead) && LeftUnderControlOf(ev, dead) == source.Controller &&
		LeftAsSubtype(ev, dead, source.TextSubtype("Knight"))
}

// s82Mom09CounterOnEachOtherMerfolkOrKnight is Vodalian Wave-Knight's body:
// one +1/+1 counter on each other Merfolk and/or Knight the controller
// controls (a Merfolk Knight gets one counter, not two).
func S82Mom09CounterOnEachOtherMerfolkOrKnight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == src.InstanceID || c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if c.HasSubtype(src.TextSubtype("Merfolk")) || c.HasSubtype(src.TextSubtype("Knight")) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// s82Mom09CastTwoOrMoreSpellsThisTurn is "Activate only if you've cast two or
// more spells this turn".
func S82Mom09CastTwoOrMoreSpellsThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).Total >= 2
}
