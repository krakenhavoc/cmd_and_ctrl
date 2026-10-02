package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_cards_test.go — ADR 0108 Delivery PR 1 (#1886, #1887):
// "if it would die this turn, exile it instead" and "can't be
// regenerated this turn", each card against the board that tells its
// wording apart.

const (
	p1LavaCoilOracle      = "fa71db44-5181-4c51-8b24-7fbedf36e3ca"
	p1DemonfireOracle     = "314a5c76-1a68-433b-a383-1834400254a8"
	p1DisintegrateOracle  = "92d6af2f-728e-4e41-87cb-5c90878a2f2f"
	p1IncinerateOracle    = "d8fd7a34-8418-4e98-b79b-119c4348c667"
	p1AngerOracle         = "3a7fe095-8278-4b1d-bec4-19b35bdcdd1b"
	p1FlayingOracle       = "cc2d016a-af44-427b-a25a-593274369449"
	p1EclipseOracle       = "5beb8d6e-d3c1-46a5-8516-d6bf66413cff"
	p1MalfunctionOracle   = "c7ecaa1a-fbf7-436b-a1c9-d7810b0dc5dc"
	p1WhippoorwillOracle  = "84050a10-e1f1-413e-aa21-5c1f47bb2a64"
	p1PreventEverything   = 20
	p1BearPower, p1BearTo = 2, 2
)

// p1Creature puts a vanilla creature on the battlefield for `owner`.
func p1Creature(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Creature", TypeLine: "Creature — Bear",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func p1Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

func p1Regenerate(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(id); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})
}

func p1ShieldCreature(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, id, p1PreventEverything, false, "shield") })
}

func p1WantZone(t *testing.T, g *game.Game, id uuid.UUID, want game.ZoneKind, what string) {
	t.Helper()
	if z := zoneOf(g, id); z != want {
		t.Errorf("%s is in %q, want %q", what, z, want)
	}
}

// Lava Coil: the creature it kills is exiled, and so is one that
// survives the damage and dies later this turn.
func TestP1LavaCoilExilesWhatItKillsAndWhatDiesLater(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	castCatalogSpell(t, g, "Lava Coil", "Sorcery", p1LavaCoilOracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear Lava Coil killed")

	big := p1Creature(g, opp, 5, 5)
	castCatalogSpell(t, g, "Lava Coil", "Sorcery", p1LavaCoilOracle, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 5/5 after 4 damage")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 5/5 destroyed later this turn")
}

// "If that creature would die" is the spell's effect: it holds even when
// the damage is prevented.
func TestP1LavaCoilMarksTheTargetEvenWhenTheDamageIsPrevented(t *testing.T) {
	g := newCatalogGame(t)
	bear := p1Creature(g, g.Seats[1].ID, p1BearPower, p1BearTo)
	p1ShieldCreature(g, bear)
	castCatalogSpell(t, g, "Lava Coil", "Sorcery", p1LavaCoilOracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, bear); got != 0 {
		t.Fatalf("the shielded bear has %d damage", got)
	}
	p1Destroy(t, g, bear)
	p1WantZone(t, g, bear, game.ZoneExile, "the shielded bear destroyed later")
}

// Demonfire: "a creature dealt damage this way". A creature whose damage
// was all prevented is not marked; one that was dealt damage is.
func TestP1DemonfireMarksOnlyACreatureItDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	shielded := p1Creature(g, opp, 5, 5)
	p1ShieldCreature(g, shielded)
	castXSpell(t, g, "Demonfire", "Sorcery", p1DemonfireOracle, "{X}{R}", 3, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature Demonfire dealt no damage")

	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	castXSpell(t, g, "Demonfire", "Sorcery", p1DemonfireOracle, "{X}{R}", 3, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear Demonfire killed")
}

// Hellbent: with no cards in hand the damage can't be prevented, so the
// shielded creature is dealt it, dies, and is exiled.
func TestP1DemonfireHellbentGetsThroughAShield(t *testing.T) {
	g := newCatalogGame(t)
	shielded := p1Creature(g, g.Seats[1].ID, p1BearPower, p1BearTo)
	p1ShieldCreature(g, shielded)
	active := g.Seats[g.Turn.ActiveSeat]
	active.Hand.Cards = nil
	id := castXSpell(t, g, "Demonfire", "Sorcery", p1DemonfireOracle, "{X}{R}", 3, pr6Card(shielded))
	var uncounterable bool
	g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("a hellbent Demonfire can be countered")
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, shielded, game.ZoneExile, "the shielded bear under a hellbent Demonfire")
}

// Disintegrate: a regenerating creature is not regenerated, and is
// exiled.
func TestP1DisintegrateStopsRegenerationAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	bear := p1Creature(g, g.Seats[1].ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, bear)
	castXSpell(t, g, "Disintegrate", "Sorcery", p1DisintegrateOracle, "{X}{R}", 2, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the regenerating bear")
}

// Incinerate: a creature dealt the damage can't be regenerated; one
// whose damage was prevented can.
func TestP1IncinerateStopsRegenerationOnlyWhenItDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	p1Regenerate(t, g, bear)
	castCatalogSpell(t, g, "Incinerate", "Instant", p1IncinerateOracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Incinerate dealt 3")

	shielded := p1Creature(g, opp, p1BearPower, p1BearTo)
	p1ShieldCreature(g, shielded)
	p1Regenerate(t, g, shielded)
	castCatalogSpell(t, g, "Incinerate", "Instant", p1IncinerateOracle, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneBattlefield, "a regenerating creature Incinerate dealt nothing")
}

// Anger of the Gods: each creature dealt damage is marked; one whose
// damage was prevented is not.
func TestP1AngerOfTheGodsMarksEachCreatureItDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	big := p1Creature(g, opp, 5, 5)
	shielded := p1Creature(g, opp, 5, 5)
	p1ShieldCreature(g, shielded)
	castCatalogSpell(t, g, "Anger of the Gods", "Sorcery", p1AngerOracle, nil)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear Anger killed")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 5/5 Anger damaged, destroyed later")
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "the shielded 5/5")
}

// Flaying Tendrils: the -2/-2 kills, the replacement exiles, and a
// creature that arrives later this turn is covered too (CR 611.2c).
func TestP1FlayingTendrilsCoversCreaturesThatArriveLater(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Flaying Tendrils", p1FlayingOracle},
		{"Malicious Malfunction", p1MalfunctionOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1].ID
			bear := p1Creature(g, opp, p1BearPower, p1BearTo)
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, bear, game.ZoneExile, "the bear")
			late := p1Creature(g, opp, 3, 3)
			p1Destroy(t, g, late)
			p1WantZone(t, g, late, game.ZoneExile, "a creature that arrived later")
		})
	}
}

// Malicious Eclipse: only a creature an opponent controls is exiled.
func TestP1MaliciousEclipseExilesOnlyOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	mine := p1Creature(g, me, p1BearPower, p1BearTo)
	theirs := p1Creature(g, opp, p1BearPower, p1BearTo)
	castCatalogSpell(t, g, "Malicious Eclipse", "Sorcery", p1EclipseOracle, nil)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, mine, game.ZoneGraveyard, "my bear")
	p1WantZone(t, g, theirs, game.ZoneExile, "the opponent's bear")
}

// Whippoorwill: the target can't be regenerated, its damage can't be
// prevented, and when it dies it is exiled — after it died, so it was a
// death.
func TestP1WhippoorwillMarksItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	bird := pushCatalogPermanent(g, me.ID, "Whippoorwill", "Creature — Bird", p1WhippoorwillOracle, false)
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	p1Regenerate(t, g, bear)
	p1ShieldCreature(g, bear)
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Targets: pr6Card(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var noRegen bool
	g.ReadSnapshot(func() { noRegen = g.PermanentCantBeRegeneratedForEffect(bear) })
	if !noRegen {
		t.Fatal("the bear can still be regenerated")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), bear, 2); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear after it died")
}
