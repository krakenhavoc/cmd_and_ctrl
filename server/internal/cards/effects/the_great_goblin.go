package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Great Goblin — Legendary Creature — Goblin Noble {1}{B/R}{B/R}, 3/2
// (The Hobbit, HOB):
//
//	"Whenever you put one or more counters on a Goblin, Orc, or Army you
//	 control, The Great Goblin deals 2 damage to target opponent.
//	 Whenever another Goblin, Orc, or Army you control dies, exile the top
//	 card of your library. You may play it until the end of your next turn."
//
// The first ability is Exemplar of Light's reading of EventCounterPlaced,
// widened to every kind of counter: the placement must have raised the
// total (a removal is not a placement, b11CountersWerePlaced), it must be
// the controller who put them (ev.Actor), and the permanent must be a
// Goblin, Orc or Army you control after the layers (the Army from Amass is
// a Goblin as well). "One or more" is one trigger per permanent per batch
// of events: two kinds of counter put on the same Goblin by one proliferate
// are one trigger (hobA9PerPermanent), while counters on two Goblins at
// once are two, as the engine's other per-permanent counter triggers are.
// The damage is dealt by The Great Goblin to the opponent chosen as the
// trigger goes on the stack.
//
// The second ability reads the dead creature's last-known characteristics
// (CR 603.10a), so a Goblin that was only a Goblin through an effect still
// counts, and a token Army does too. The exile is Prosper's: the card may be
// played, a land included, until the end of its controller's next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83dcfac0-6efd-4e37-9402-15f9889e84e1",
		Name:         "The Great Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(hobA9PerPermanent(On(game.EventCounterPlaced, hobA9CountersOnYourGoblinOrcOrArmy,
				"The Great Goblin — 2 damage to target opponent", sourceDealsDamageToEachLegalTarget(2))),
				TargetPlayer("target opponent", Opponent())),
			On(game.EventLTB, hobA9AnotherGoblinOrcOrArmyYouControlDied,
				"The Great Goblin — exile the top card of your library; you may play it until the end of your next turn",
				b20ExileTopUntilEndOfNextTurn),
		},
	})
}

// hobA9GoblinOrcOrArmy are the creature types of The Great Goblin's text.
var hobA9GoblinOrcOrArmy = [...]string{"Goblin", "Orc", "Army"}

// hobA9CountersOnYourGoblinOrcOrArmy is the first ability's condition: the
// source's controller put counters on a Goblin, Orc or Army they control.
func hobA9CountersOnYourGoblinOrcOrArmy(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b11CountersWerePlaced(ev, ev.Target, ev.Label, g) || b11CounterPlacerOf(ev, g) != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || c.Controller != source.Controller {
		return false
	}
	for _, t := range hobA9GoblinOrcOrArmy {
		if c.HasSubtype(source.TextSubtype(t)) {
			return true
		}
	}
	return false
}

// hobA9AnotherGoblinOrcOrArmyYouControlDied is the second ability's
// condition, read off the dead creature's last-known characteristics.
func hobA9AnotherGoblinOrcOrArmyYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	for _, t := range hobA9GoblinOrcOrArmy {
		if leftAsSubtype(ev, dead, source.TextSubtype(t)) {
			return true
		}
	}
	return false
}

// hobA9PerPermanent collapses a trigger's events to one per permanent per
// batch (OncePerBatch keyed on the event's target): the "one or more
// counters on a <permanent>" reading when a single placement is reported
// once per kind of counter.
func hobA9PerPermanent(t game.TriggeredAbility) game.TriggeredAbility {
	t.OncePerBatch = true
	t.BatchKey = hobA9EventTarget
	return t
}

// hobA9EventTarget is the batch dimension of hobA9PerPermanent: the
// permanent the event names.
func hobA9EventTarget(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Target.String() }
