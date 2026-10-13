package effects

import "testing"

const dwarvenLightsmithOracle = "cb9210eb-f5cb-430d-b893-b1d814117043"

// TestDwarvenLightsmithPumpsEachCreatureYouControl: "your team" is just
// the caster, so the ETB gives +1/+1 to every creature the caster
// controls (the Lightsmith included) and to nothing an opponent
// controls.
func TestDwarvenLightsmithPumpsEachCreatureYouControl(t *testing.T) {
	assistCardSpec(t, dwarvenLightsmithOracle, "Dwarven Lightsmith")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	foe := pushVanillaCreature(g, opp.ID, "Foe", 2, 2)

	id := castCatalogSpell(t, g, "Dwarven Lightsmith", "Creature — Dwarf Cleric", dwarvenLightsmithOracle, nil)
	passPriorityAroundTable(t, g)

	if p := apaLive(g, bear).CurrentPower(); p != 3 {
		t.Errorf("Bear power %d, want 3", p)
	}
	if p := apaLive(g, id).CurrentPower(); p != apaLive(g, id).Power+1 {
		t.Errorf("Lightsmith power %d, want its printed power plus 1", p)
	}
	if p := apaLive(g, foe).CurrentPower(); p != 2 {
		t.Errorf("opponent's creature power %d, want 2 (not pumped)", p)
	}
}
