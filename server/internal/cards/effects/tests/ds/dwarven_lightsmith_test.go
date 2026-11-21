package testsds

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const dwarvenLightsmithOracle = "cb9210eb-f5cb-430d-b893-b1d814117043"

// TestDwarvenLightsmithPumpsEachCreatureYouControl: "your team" is just
// the caster, so the ETB gives +1/+1 to every creature the caster
// controls (the Lightsmith included) and to nothing an opponent
// controls.
func TestDwarvenLightsmithPumpsEachCreatureYouControl(t *testing.T) {
	effectstest.AssistCardSpec(t, dwarvenLightsmithOracle, "Dwarven Lightsmith")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	foe := effectstest.PushVanillaCreature(g, opp.ID, "Foe", 2, 2)

	id := effectstest.CastCatalogSpell(t, g, "Dwarven Lightsmith", "Creature — Dwarf Cleric", dwarvenLightsmithOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if p := effectstest.ApaLive(g, bear).CurrentPower(); p != 3 {
		t.Errorf("Bear power %d, want 3", p)
	}
	if p := effectstest.ApaLive(g, id).CurrentPower(); p != effectstest.ApaLive(g, id).Power+1 {
		t.Errorf("Lightsmith power %d, want its printed power plus 1", p)
	}
	if p := effectstest.ApaLive(g, foe).CurrentPower(); p != 2 {
		t.Errorf("opponent's creature power %d, want 2 (not pumped)", p)
	}
}
