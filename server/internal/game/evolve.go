package game

// evolve.go — CR 702.100, the second KEYWORD TRIGGER the engine derives
// from an object's ability list rather than from its catalog entry
// (#1805, ADR 0106 §3). Prowess (prowess.go) is the model.
//
//	702.100a Evolve is a triggered ability. "Evolve" means "Whenever a
//	         creature you control enters, if that creature's power is
//	         greater than this creature's power and/or that creature's
//	         toughness is greater than this creature's toughness, put a
//	         +1/+1 counter on this creature."
//	702.100b A creature "evolves" when one or more +1/+1 counters are
//	         put on it as a result of its evolve ability resolving.
//	702.100c A creature can't have a greater power or toughness than a
//	         noncreature permanent.
//	702.100d If a creature has multiple instances of evolve, each
//	         triggers separately.
//
// WHY A TOKEN. The same reason prowess is one: evolve has no parameter,
// and it is GRANTED (Tyranid Prime's "other creatures you control have
// evolve", Propagator Drone's "creature tokens you control have
// evolve"). A token in Characteristic.Abilities reaches a printed
// evolve through the deck importer, a token template's Keywords and a
// layer-6 grant with no catalog entry at all, so the seven creatures
// whose only text is evolve and other canonical keywords (Cloudfin
// Raptor, Shambleshark, …) need no card file.
//
// THE INTERVENING IF (CR 603.4). The comparison is made twice: as the
// creature enters (AppliesTo — no comparison, no trigger) and again as
// the trigger resolves (resolveEvolve — no comparison, no counter).
// Each side is read the way CR 608.2h reads it at resolution: the
// entered creature through the item's carried trigger object, live
// while it is on the battlefield and as it last existed once it has
// left (PermanentForEffect); the evolving creature live, and not at
// all once it has left or come back as a new object (CR 400.7), since
// a counter can't be put on last-known information.
//
// "EVOLVES" (CR 702.100b). The counter goes through the ordinary CR 614
// counter window, so Hardened Scales, Doubling Season and "can't have
// counters" all apply, and it is sequenced behind the placement
// (AddCounterByThenForEffect): EventEvolved is emitted only when at
// least one counter actually landed, which is what "whenever this
// creature evolves" (Renegade Krasis, Watchful Radstag) watches.
//
// ORDERING (CR 603.3b). Evolve triggers do NOT commute the way prowess
// triggers do (#1511): each one changes the power and toughness the
// next one's condition reads, so the order is a real choice and the
// prompt is kept. The label names the creature that entered, so two
// evolve triggers on one creature from two different entering
// creatures (CR 603.6a — they entered together and each saw the
// other) are told apart by seatNeedsTriggerOrder rather than ordered
// silently.

import "github.com/google/uuid"

// KeywordEvolve is CR 702.100's canonical token.
const KeywordEvolve = "evolve"

// evolveLabelPrefix starts every evolve trigger's stack label.
const evolveLabelPrefix = "Evolve"

// evolveTrigger is the one TriggeredAbility a single instance of evolve
// is. A package-level value, never mutated: keywordTriggersFor hands out
// copies of it.
var evolveTrigger = TriggeredAbility{
	Keyword: KeywordEvolve,
	Watches: []EventKind{EventETB},
	// "Whenever a creature you control enters, if that creature's power
	// is greater … and/or … toughness is greater". The ETB harvest has
	// already caught the layer cache up (triggerHarvester.OnEvent), so
	// both creatures are read with every continuous effect and every
	// counter they have, including counters the entering creature
	// entered with (CR 614.1c).
	AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
		if ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
			// A creature is never greater than itself.
			return false
		}
		entered := findBattlefieldCard(g, ev.CardID)
		if entered == nil || entered.Controller != source.Controller {
			return false
		}
		return evolveComparisonHolds(
			entered.IsCreature(), entered.PowerForComparison(), entered.CurrentToughness(),
			source.IsCreature(), source.PowerForComparison(), source.CurrentToughness(),
		)
	},
	Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
		// Keyed directly, like prowess: evolve has no catalog row, and a
		// table with one waiting on the stack is still a restore point.
		// The entered creature rides on item.Trigger, which the
		// harvester stamps after Build (stampTriggerContext).
		return NewKeyedTriggeredItem(source, evolveLabelFor(g, ev.CardID), evolveGrowBody, EffectParams{})
	},
}

// evolveLabelFor is the stack label: "Evolve — <entered creature>". The
// name is the entering object's EFFECTIVE name, so a face-down creature
// (CR 708.2, no name) is never named by its card.
func evolveLabelFor(g *Game, enteredID uuid.UUID) string {
	if c := findBattlefieldCard(g, enteredID); c != nil {
		if name := c.Effective().Name; name != "" && !c.FaceDown {
			return evolveLabelPrefix + " — " + name + " entered"
		}
	}
	return evolveLabelPrefix + " — a creature entered"
}

// evolveComparisonHolds is CR 702.100a's condition, with CR 702.100c:
// the entered creature's power is greater than the evolving creature's
// power, or its toughness is greater than the evolving creature's
// toughness. Power is compared with power and toughness with toughness,
// and either side not being a creature makes it false.
func evolveComparisonHolds(enteredIsCreature bool, enteredPower, enteredToughness int, selfIsCreature bool, selfPower, selfToughness int) bool {
	if !enteredIsCreature || !selfIsCreature {
		return false
	}
	return enteredPower > selfPower || enteredToughness > selfToughness
}

// evolveGrowBody is registered under "evolve/grow" (ADR 0106 §3,
// ADR 0041 P9): an on-disk identity, never renamed. An older binary
// refuses a restore point naming it with ErrUnknownEffectKey, which is
// the designed rollback case.
var evolveGrowBody = SimpleDelayedBody("evolve/grow", resolveEvolve)

// resolveEvolve re-checks the comparison (CR 603.4) and, if it still
// holds, puts one +1/+1 counter on the evolving creature, then emits
// EventEvolved if a counter landed (CR 702.100b). A package-level func,
// so it captures nothing and survives Clone.
func resolveEvolve(g *Game, item *StackItem) error {
	if g.AbilitySourceIsNewObjectForEffect(item) {
		return nil
	}
	self := findBattlefieldCard(g, item.SourceCardID)
	if self == nil || item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	entered, ok := g.PermanentForEffect(item.Trigger.Object.Ref())
	if !ok {
		// The entered creature left by a route that keeps no record of
		// it (its owner left the game): the condition can't be checked,
		// so it isn't true.
		return nil
	}
	// PermanentForEffect caught the layer cache up, so the evolving
	// creature is read current too — a pump in response counts.
	self = findBattlefieldCard(g, item.SourceCardID)
	if self == nil || !evolveComparisonHolds(
		typeListHas(entered.Characteristic.Types, "creature"), entered.Power, entered.Toughness,
		self.IsCreature(), self.PowerForComparison(), self.CurrentToughness(),
	) {
		return nil
	}
	selfID, controller := self.InstanceID, item.Controller
	return g.AddCounterByThenForEffect(controller, selfID, CounterPlusOne, 1, func(g *Game, placed int) error {
		if placed <= 0 {
			return nil
		}
		g.EmitEvent(Event{
			Kind:   EventEvolved,
			CardID: selfID,
			Target: selfID,
			Source: selfID,
			Actor:  controller,
			Amount: placed,
		})
		return nil
	})
}

// EvolveCount reports how many instances of evolve the card has — a
// read for tests, the bot and the view, through the same walk the
// harvester uses.
func EvolveCount(c *Card) int {
	return countAbilityTokens(c, KeywordEvolve)
}
