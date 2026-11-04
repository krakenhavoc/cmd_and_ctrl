package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kraven the Hunter — Legendary Creature — Human Warrior Villain {1}{B}{G},
// 4/3 (Marvel's Spider-Man, #4460):
//
//	"Trample
//	 Whenever a creature an opponent controls with the greatest power among
//	 creatures that player controls dies, draw a card and put a +1/+1
//	 counter on Kraven the Hunter."
//
// A leaves-the-battlefield trigger looks back in time (CR 603.10a): the dead
// creature's power is its last-known power (the departure record the exit
// wrote), and "the greatest power among creatures that player controls" is
// judged over the creatures that player controlled immediately before the
// event, so every creature of theirs that left in the same event batch
// counts as well as the ones still on the battlefield. A tie for the
// greatest power qualifies each tied creature, so a board wipe over two
// equal giants fires once for each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3e2bfa3a-ae83-453a-8e3f-ca6205a9af12",
		Name:            "Kraven the Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, kravenTheHunterGreatestPowerOpponentCreatureDied,
				"Kraven the Hunter — draw a card and put a +1/+1 counter on Kraven",
				Do(DrawCards{N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// kravenTheHunterGreatestPowerOpponentCreatureDied is the trigger condition:
// a creature an opponent controlled died, and no other creature that player
// controlled just before the event had greater power.
func kravenTheHunterGreatestPowerOpponentCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	owner := leftUnderControlOf(ev, dead)
	if owner == source.Controller {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(ev.CardID)
	if !ok {
		return false
	}
	power := info.Power
	// Creatures of theirs still on the battlefield.
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.IsCreature() && c.PowerForComparison() > power {
			return false
		}
	}
	// Creatures of theirs that left in the same event batch (they were on
	// the battlefield when the event happened).
	if ev.Batch == 0 {
		return true
	}
	seen := map[uuid.UUID]bool{ev.CardID: true}
	for _, e := range g.EventsThisTurn() {
		if e.Kind != game.EventLTB || e.Batch != ev.Batch || e.CardID == uuid.Nil || seen[e.CardID] {
			continue
		}
		seen[e.CardID] = true
		other, found := g.LookupCardForEffect(e.CardID)
		if !found || !leftAsType(e, other, "creature") || leftUnderControlOf(e, other) != owner {
			continue
		}
		if rec, has := g.LastKnownPermanentForEffect(e.CardID); has && rec.Power > power {
			return false
		}
	}
	return true
}
