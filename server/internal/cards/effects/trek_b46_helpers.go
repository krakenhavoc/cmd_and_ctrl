package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b46_helpers.go — shared pieces of Star Trek Commander batch 46
// (#4974).

// trekB46FaerieToken is the 1/1 blue Faerie creature token with flying
// that Worf, Still Not Merry; La Forge, Thoughtful as Ever; and Data,
// Professor of Humanities create.
const trekB46FaerieToken = "1/1 blue Faerie with flying"

// onceEachTurnMayDrawThatMany is "Whenever you put one or more +1/+1
// counters on <a creature `match` admits>, you may draw that many cards.
// Do this only once each turn." (Terrasymbiosis; Chekov and McCoy).
//
// The trigger is WheneverYouPutCountersOfKindOnACreature: once per
// creature per placement (CR 603.2c), +1/+1 counters only, put by you.
// "That many" is the +1/+1 counters that placement put on the creature,
// after replacements, read as the trigger resolves.
//
// "Do this only once each turn" limits the draw, not the trigger:
// every placement triggers it and asks its "you may" as it resolves; the
// first "yes" draws and uses up the turn, and a "no" does not. The record
// is per object (NoteDoneThisTurn) under doneLabel, and the gate is also
// read as the ability triggers, so no trigger is made once the turn's
// draw is done.
func onceEachTurnMayDrawThatMany(match func(g *game.Game, source *game.Card, c game.Card) bool, label, doneLabel, question string) game.TriggeredAbility {
	t := WheneverYouPutCountersOfKindOnACreature(game.CounterPlusOne, match, label,
		func(g *game.Game, item *game.StackItem) error {
			if b28DoneThisTurn(g, item.SourceCardID, doneLabel) {
				return nil
			}
			n := plusOneCountersPutInTheTriggeringBatch(g, item)
			if n <= 0 {
				return nil
			}
			b28NoteDoneThisTurn(g, item.SourceCardID, doneLabel)
			return DrawCards{N: n}.Apply(NewContext(g, item))
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return when(ev, source, ch, g) && !b28DoneThisTurn(g, source.InstanceID, doneLabel)
	}
	return OptionalDrawThatMany(t, question, plusOneCountersPutInTheTriggeringBatch)
}

// plusOneCountersPutInTheTriggeringBatch is how many +1/+1 counters the
// trigger's controller put in the placement that triggered it: "that
// many" (#4016, OptionalDrawThatMany).
func plusOneCountersPutInTheTriggeringBatch(g *game.Game, item *game.StackItem) int {
	if item.Trigger == nil {
		return 0
	}
	return CountersOfKindYouPutInTheSameBatch(g, item.Trigger.Event, item.Controller, game.CounterPlusOne)
}

// trekB46IsThis is the match for "on <this creature>": the creature the
// counters went on is the trigger's source itself.
func trekB46IsThis(_ *game.Game, source *game.Card, c game.Card) bool {
	return source != nil && c.InstanceID == source.InstanceID
}

// trekB46NontokenPermanentsYouControlThatEnteredThisTurn counts Spock,
// Voice of Reason's "nontoken permanent you control that entered this
// turn", read as the ability resolves (CR 608.2h). The entry tally is per
// object, so a permanent that left and came back is counted once, as the
// new object it is.
//
// Caller must hold g.mu.
func trekB46NontokenPermanentsYouControlThatEnteredThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && !IsToken(c) && g.EnteredThisTurn(c.InstanceID) {
			n++
		}
	}
	return n
}

// trekB46PermanentsYouControlYouPutACounterOnThisTurn counts La Forge,
// Thoughtful as Ever's "permanent you control that you've put a counter
// on this turn": a counter of any kind that you put on that very object
// this turn (socB25YouPutACounterOnThisObjectThisTurn), whether or not
// it is still there, on a permanent you control as the ability resolves.
//
// Caller must hold g.mu.
func trekB46PermanentsYouControlYouPutACounterOnThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && socB25YouPutACounterOnThisObjectThisTurn(g, you, c.InstanceID) {
			n++
		}
	}
	return n
}

// trekB46AttackingNonFaerieCreatures counts Worf, Still Not Merry's
// "attacking non-Faerie creature", every player's, through the layered
// view (a changeling is a Faerie), as the ability resolves. The word
// "Faerie" is read through the source's text (CR 612.2); a nil source
// reads it as printed.
//
// Caller must hold g.mu.
func trekB46AttackingNonFaerieCreatures(g *game.Game, source *game.Card) int {
	faerie := source.TextSubtype("Faerie")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil && !c.HasSubtype(faerie) {
			n++
		}
	}
	return n
}

// crewsAndStationsAsThoughPowerTripled is Sulu, Heroic Helmsman's "Sulu
// stations Spacecraft and crews Vehicles as though his power were
// tripled": CrewsAsThoughPowerGreater's layer 6 self-only shape, with a
// Multiplier of 3 for crew (CR 702.122a) and for the station ability of a
// Spacecraft (CR 702.184a and c). It changes only what the creature counts
// for when it pays those costs (Card.CountedPower), never its power, and a
// creature that loses all abilities stops counting extra (CR 613.1f).
func crewsAndStationsAsThoughPowerTripled() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		Label:     "This creature stations Spacecraft and crews Vehicles as though its power were tripled.",
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.CostPower = append(c.CostPower, game.CostPowerRule{Uses: game.CostPowerCrew | game.CostPowerStation, Multiplier: 3})
		},
	}
}

// trekB46NonHumanCreatureEntered is Uhura, Xenolinguist's "whenever a
// non-Human creature enters": any player's, token or card, judged on the
// battlefield as it entered. "Human" is read through the source's text
// (CR 612.2).
func trekB46NonHumanCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b33AnyCreatureEntered(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.HasSubtype(source.TextSubtype("Human"))
}
