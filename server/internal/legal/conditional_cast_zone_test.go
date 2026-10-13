package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// conditional_cast_zone_test.go — #3105. A card that opens a cast zone
// only while a condition holds (Spec.CastableZonesIf, ADR 0066's
// 2026-10-13 amendment) is offered to a bot exactly while the engine
// would accept the cast — the enumerator asks the same predicate
// (Game.CardCastableFromAnyFaceLocked). A bot picks from this list, so
// "offered / not offered" here is "the bot can / cannot".

const (
	oracleLightwheelEnhancements = "17fae94e-558f-449a-b30c-e3c37a068b79" // max speed: graveyard
	oracleGravecrawler           = "09ff28b1-b6c9-48e6-b12e-2f0e644f709f" // a Zombie: graveyard
)

func TestEnumeratorOffersLightwheelOnlyAtMaxSpeed(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Plains")
	battlefieldCard(g, seat, creature("Bear", "{1}{G}", 2, 2))

	aura := graveyardCard(seat, game.Card{
		Name: "Lightwheel Enhancements", TypeLine: "Enchantment — Aura", ManaCost: "{W}",
		OracleID: oracleLightwheelEnhancements,
	})

	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), aura)); n != 0 {
		t.Errorf("offered %d casts from the graveyard with no speed", n)
	}
	g.SetSpeedForTest(seat.ID, game.MaxSpeed-1)
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), aura)); n != 0 {
		t.Errorf("offered %d casts from the graveyard at speed %d", n, game.MaxSpeed-1)
	}

	g.SetSpeedForTest(seat.ID, game.MaxSpeed)
	moves := legal.EnumerateFor(g, seat.ID)
	if n := len(castMovesFor(moves, aura)); n == 0 {
		t.Fatalf("not offered at max speed: %v", labels(moves))
	}
	if !hasLabel(moves, "Cast Lightwheel Enhancements from graveyard") {
		t.Errorf("label does not name the graveyard: %v", labels(moves))
	}
	// The contract: every offered move is one the dispatcher accepts.
	dispatchAll(t, g, seat.ID, moves)

	// The button goes again when the speed does.
	g.SetSpeedForTest(seat.ID, 2)
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), aura)); n != 0 {
		t.Errorf("still offered %d casts after the speed fell", n)
	}
}

func TestEnumeratorOffersGravecrawlerOnlyWhileYouControlAZombie(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Swamp")

	crawler := graveyardCard(seat, game.Card{
		Name: "Gravecrawler", TypeLine: "Creature — Zombie", ManaCost: "{B}",
		Power: 2, Toughness: 1, OracleID: oracleGravecrawler,
	})
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), crawler)); n != 0 {
		t.Errorf("offered %d casts with no Zombie on the battlefield", n)
	}

	battlefieldCard(g, seat, game.Card{
		Name: "Walking Corpse", TypeLine: "Creature — Zombie", ManaCost: "{1}{B}", Power: 2, Toughness: 2,
	})
	moves := legal.EnumerateFor(g, seat.ID)
	if n := len(castMovesFor(moves, crawler)); n == 0 {
		t.Fatalf("not offered with a Zombie in play: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A card that declares nothing is untouched: a vanilla creature in the
// graveyard is never offered, at any speed.
func TestEnumeratorStillRefusesAnUndeclaredGraveyardCardAtMaxSpeed(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Forest")
	g.SetSpeedForTest(seat.ID, game.MaxSpeed)

	bear := graveyardCard(seat, creature("Bear", "{1}{G}", 2, 2))
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), bear)); n != 0 {
		t.Errorf("offered %d casts of an undeclared graveyard card", n)
	}
}
