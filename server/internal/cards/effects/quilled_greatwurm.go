package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quilled Greatwurm — Creature — Wurm {4}{G}{G}, 7/7 (Foundations,
// #4465, seam #5390):
//
//	"Trample
//	 Whenever a creature you control deals combat damage during your
//	 turn, put that many +1/+1 counters on it. (It must survive to get
//	 the counters.)
//	 You may cast this card from your graveyard by removing six counters
//	 from among creatures you control in addition to paying its other
//	 costs."
//
// THE TRIGGER is once per creature per combat damage step, for the
// whole of what that creature dealt then (CR 510.2: combat damage is
// dealt simultaneously): a trampler that puts 3 into a blocker and 4
// into a player gets one trigger for seven, not two triggers. That is
// game.BatchTotalDamageDealt, the per-source total Spinneret and
// Spiderling reads, with no threshold. "That many" is the damage
// actually dealt, so prevented damage grows nothing. Damage to a
// blocker, a player, a planeswalker or a battle all count: the text
// names no recipient. "During your turn" is the active-player check, so
// a creature you control blocking on an opponent's turn grows nothing.
//
// "It" is the creature that dealt the damage (the trigger's subject,
// SubjectIsEventSource), and the reminder text is CR 400.7: a creature
// that died to the same damage, or left and came back as a new object
// before the trigger resolves, gets nothing.
//
// THE GRAVEYARD CAST is CastFromGraveyardByRemovingCounters: the printed
// {4}{G}{G}, plus six counters of any kinds from among creatures you
// control, split as you choose (CR 601.2b, 601.2h), and refused when
// they hold fewer than six in all (CR 118.3). Normal timing applies (the
// card's ruling), and there is no exile clause, so the Greatwurm can be
// cast from the graveyard again each time it gets back there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "01635d21-aeda-4754-9d94-68fcd0cd6556",
		Name:             "Quilled Greatwurm",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{CastFromGraveyardByRemovingCounters("{4}{G}{G}", 6, "creatures you control", Creature())},
		Triggered: []game.TriggeredAbility{{
			Watches:              []game.EventKind{game.EventDealDamage},
			AppliesTo:            quilledGreatwurmCombatDamageOnYourTurn,
			AtBatchEnd:           ThatMuchDamageDealtAtLeast(1),
			SubjectIsEventSource: true,
			Key:                  "Quilled Greatwurm — put that many +1/+1 counters on it",
			Effect:               quilledGreatwurmGrow,
		}},
	})
}

// quilledGreatwurmCombatDamageOnYourTurn is one slice of the combat
// damage a creature you control dealt on your turn. The dealer's
// controller is the one the event captured as the damage was dealt
// (Event.Actor), so a creature that died to the same damage still
// counts as yours.
func quilledGreatwurmCombatDamageOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if ev.Actor != source.Controller || !isActivePlayer(g, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsCreature()
}

// quilledGreatwurmGrow puts the settled total on the creature that dealt
// it, if it is still on the battlefield as the same object.
func quilledGreatwurmGrow(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.Source
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if o := item.Trigger.Object; o != nil && o.ID == id && g.ObjectEpochForEffect(id) != o.Epoch {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: settledEventAmount(item)}.Apply(NewContext(g, item))
}
