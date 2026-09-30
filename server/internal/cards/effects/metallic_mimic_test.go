package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const metallicMimicOracle = "1f91297a-ec2b-4ea8-9198-aa1daac20ff8"

// TestMetallicMimicGivesOtherChosenTypeCreaturesAnExtraCounter —
// naming Elf, another Elf entering under the same controller gets an
// additional +1/+1 counter; a non-Elf creature does not.
func TestMetallicMimicGivesOtherChosenTypeCreaturesAnExtraCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNamedTribePermanent(t, g, me.ID, "Metallic Mimic", "Artifact Creature — Shapeshifter", metallicMimicOracle, "Elf")

	elfID := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: elfID, Name: "Elvish Bear", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, elfID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Elvish Bear: %v", err)
	}
	passPriorityAroundTable(t, g)

	var counters int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == elfID {
				counters = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if counters != 1 {
		t.Errorf("Elf entering with Mimic naming Elf: counters = %d, want 1", counters)
	}

	nonElfID := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: nonElfID, Name: "Plain Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, nonElfID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Plain Bear: %v", err)
	}
	passPriorityAroundTable(t, g)

	var nonElfCounters int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == nonElfID {
				nonElfCounters = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if nonElfCounters != 0 {
		t.Errorf("a non-Elf creature should get no bonus counter: %d", nonElfCounters)
	}
}
