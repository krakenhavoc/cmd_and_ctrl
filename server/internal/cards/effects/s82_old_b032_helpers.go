package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b032_helpers.go — shared pieces of OLD batch 32 (#5501, Modern
// Horizons 2).

// OldB032GainLifePerDiscardThisTurn is Dihada's Ploy's rider, "you gain life
// equal to the number of cards you've discarded this turn", as the
// continuation of its discard (keyed, so a table waiting on the discard is a
// restore point). The tally already holds the card just discarded. An
// on-disk key: never renamed, never reused.
var OldB032GainLifePerDiscardThisTurn = game.RegisterDiscardThen("choose-cards/old-b032-dihadas-ploy-life",
	func(g *game.Game, r game.DiscardsDone) error {
		n := g.TurnTallyFor(r.Player).CardsDiscarded
		if n <= 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(r.Source, r.Player, n)
	})

// OldB032CalibratedBlastBody is Calibrated Blast's reflexive half: damage
// equal to the revealed nonland card's mana value (Params.Amount, fixed when
// the card was revealed) to the any-target the trigger chose. It is Breeches'
// blast, under its own registered key.
var OldB032CalibratedBlastBody = game.ReflexiveBody("calibrated-blast/damage", BreechesBlastEffect, ConstTargets(TargetAny))

// OldB032RoadSearch is Road's body: "Search your library for a basic land
// card, put it onto the battlefield tapped, then shuffle."
func OldB032RoadSearch(_ *game.StackItem, ctx *Context) error {
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		TappedOnEntry: true,
		Limit:         1,
		Shuffle:       true,
		Reason:        "Road — a basic land card, onto the battlefield tapped",
	}.Apply(ctx)
}

// OldB032NontokenCreatureNotCastFromHand is Chainer's condition: a nontoken
// creature under the source's controller's control (Chainer itself included)
// entered, and its controller did not cast it from their hand. A creature
// that was not cast at all has no cast provenance and qualifies (CR 400.7d).
func OldB032NontokenCreatureNotCastFromHand(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() || c.IsToken() {
		return false
	}
	return !(c.Provenance.Caster == c.Controller && c.Provenance.FromZone == game.ZoneHand)
}

// OldB032EnteredCreatureGainsHaste gives the creature the trigger's event
// named haste until its controller's next turn.
func OldB032EnteredCreatureGainsHaste(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !OnBattlefield(g, id) {
		return nil
	}
	ctx := NewContext(g, item)
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Chainer, Nightmare Adept — haste until your next turn",
	}.Apply(ctx)
}
