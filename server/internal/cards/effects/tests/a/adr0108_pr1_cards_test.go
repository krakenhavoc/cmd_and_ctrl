package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lava Coil: the creature it kills is exiled, and so is one that
// survives the damage and dies later this turn.
func TestP1LavaCoilExilesWhatItKillsAndWhatDiesLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastCatalogSpell(t, g, "Lava Coil", "Sorcery", effectstest.P1LavaCoilOracle, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear Lava Coil killed")

	big := effectstest.P1Creature(g, opp, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Lava Coil", "Sorcery", effectstest.P1LavaCoilOracle, effectstest.Pr6Card(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, big, game.ZoneBattlefield, "the 5/5 after 4 damage")
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 5/5 destroyed later this turn")
}

// "If that creature would die" is the spell's effect: it holds even when
// the damage is prevented.
func TestP1LavaCoilMarksTheTargetEvenWhenTheDamageIsPrevented(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.P1Creature(g, g.Seats[1].ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1ShieldCreature(g, bear)
	effectstest.CastCatalogSpell(t, g, "Lava Coil", "Sorcery", effectstest.P1LavaCoilOracle, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Fatalf("the shielded bear has %d damage", got)
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the shielded bear destroyed later")
}

// Demonfire: "a creature dealt damage this way". A creature whose damage
// was all prevented is not marked; one that was dealt damage is.
func TestP1DemonfireMarksOnlyACreatureItDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	shielded := effectstest.P1Creature(g, opp, 5, 5)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.CastXSpell(t, g, "Demonfire", "Sorcery", effectstest.P1DemonfireOracle, "{X}{R}", 3, effectstest.Pr6Card(shielded))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature Demonfire dealt no damage")

	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastXSpell(t, g, "Demonfire", "Sorcery", effectstest.P1DemonfireOracle, "{X}{R}", 3, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear Demonfire killed")
}

// Hellbent: with no cards in hand the damage can't be prevented, so the
// shielded creature is dealt it, dies, and is exiled.
func TestP1DemonfireHellbentGetsThroughAShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	shielded := effectstest.P1Creature(g, g.Seats[1].ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1ShieldCreature(g, shielded)
	active := g.Seats[g.Turn.ActiveSeat]
	active.Hand.Cards = nil
	id := effectstest.CastXSpell(t, g, "Demonfire", "Sorcery", effectstest.P1DemonfireOracle, "{X}{R}", 3, effectstest.Pr6Card(shielded))
	var uncounterable bool
	g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("a hellbent Demonfire can be countered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, shielded, game.ZoneExile, "the shielded bear under a hellbent Demonfire")
}

// Disintegrate: a regenerating creature is not regenerated, and is
// exiled.
func TestP1DisintegrateStopsRegenerationAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.P1Creature(g, g.Seats[1].ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, bear)
	effectstest.CastXSpell(t, g, "Disintegrate", "Sorcery", effectstest.P1DisintegrateOracle, "{X}{R}", 2, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the regenerating bear")
}

// Incinerate: a creature dealt the damage can't be regenerated; one
// whose damage was prevented can.
func TestP1IncinerateStopsRegenerationOnlyWhenItDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, bear)
	effectstest.CastCatalogSpell(t, g, "Incinerate", "Instant", effectstest.P1IncinerateOracle, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Incinerate dealt 3")

	shielded := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.P1Regenerate(t, g, shielded)
	effectstest.CastCatalogSpell(t, g, "Incinerate", "Instant", effectstest.P1IncinerateOracle, effectstest.Pr6Card(shielded))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneBattlefield, "a regenerating creature Incinerate dealt nothing")
}

// Anger of the Gods: each creature dealt damage is marked; one whose
// damage was prevented is not.
func TestP1AngerOfTheGodsMarksEachCreatureItDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	big := effectstest.P1Creature(g, opp, 5, 5)
	shielded := effectstest.P1Creature(g, opp, 5, 5)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.CastCatalogSpell(t, g, "Anger of the Gods", "Sorcery", effectstest.P1AngerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear Anger killed")
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 5/5 Anger damaged, destroyed later")
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "the shielded 5/5")
}

// Flaying Tendrils: the -2/-2 kills, the replacement exiles, and a
// creature that arrives later this turn is covered too (CR 611.2c).
func TestP1FlayingTendrilsCoversCreaturesThatArriveLater(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Flaying Tendrils", effectstest.P1FlayingOracle},
		{"Malicious Malfunction", effectstest.P1MalfunctionOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1].ID
			bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear")
			late := effectstest.P1Creature(g, opp, 3, 3)
			effectstest.P1Destroy(t, g, late)
			effectstest.P1WantZone(t, g, late, game.ZoneExile, "a creature that arrived later")
		})
	}
}

// Malicious Eclipse: only a creature an opponent controls is exiled.
func TestP1MaliciousEclipseExilesOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	mine := effectstest.P1Creature(g, me, effectstest.P1BearPower, effectstest.P1BearTo)
	theirs := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastCatalogSpell(t, g, "Malicious Eclipse", "Sorcery", effectstest.P1EclipseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, mine, game.ZoneGraveyard, "my bear")
	effectstest.P1WantZone(t, g, theirs, game.ZoneExile, "the opponent's bear")
}

// Whippoorwill: the target can't be regenerated, its damage can't be
// prevented, and when it dies it is exiled — after it died, so it was a
// death.
func TestP1WhippoorwillMarksItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	bird := effectstest.PushCatalogPermanent(g, me.ID, "Whippoorwill", "Creature — Bird", effectstest.P1WhippoorwillOracle, false)
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, bear)
	effectstest.P1ShieldCreature(g, bear)
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear after it died")
}
