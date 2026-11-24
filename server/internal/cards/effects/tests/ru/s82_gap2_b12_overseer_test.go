package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_overseer_test.go — S82 GAP2 batch 12: Overseer of Vault 76,
// through the real enters and combat triggers, the split prompt and the
// reflexive trigger.

const gap2B12Overseer = "84f32a1d-3861-418e-a3c2-7e731a2cd55c"

func TestGap2B12OverseerIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Overseer)
	if !ok {
		t.Fatal("Overseer of Vault 76 is not registered")
	}
	if spec.Name != "Overseer of Vault 76" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// gap2B12CastCreature casts a plain creature card from the active player's
// hand and resolves it.
func gap2B12CastCreature(t *testing.T, g *game.Game, name, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Owner: me.ID, Controller: me.ID, Power: p, Toughness: tough})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func gap2B12Quest(g *game.Game, id uuid.UUID) int {
	c, _ := g.LookupCardForEffect(id)
	return c.Counters["quest"]
}

func TestGap2B12OverseerFirstContact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	overseer := gap2B12CastCreature(t, g, "Overseer of Vault 76", gap2B12Overseer, 3, 3)
	if n := gap2B12Quest(g, overseer); n != 1 {
		t.Fatalf("%d quest counters after it entered, want 1", n)
	}
	gap2B12CastCreature(t, g, "Small Friend", "", 3, 1)
	if n := gap2B12Quest(g, overseer); n != 2 {
		t.Errorf("%d quest counters after a power-3 creature entered, want 2", n)
	}
	gap2B12CastCreature(t, g, "Big Friend", "", 4, 4)
	if n := gap2B12Quest(g, overseer); n != 2 {
		t.Errorf("%d quest counters after a power-4 creature entered, want still 2", n)
	}
}

func TestGap2B12OverseerSplitsThreeQuestCountersForTheBoost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	overseer := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Overseer of Vault 76", TypeLine: "Legendary Creature — Human Advisor",
		OracleID: gap2B12Overseer, Power: 3, Toughness: 3, Counters: map[string]int{"quest": 2}})
	quest := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Quest Board", TypeLine: "Enchantment", Counters: map[string]int{"quest": 2}})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	p := effectstest.AnswerNumber(t, g, me.ID, 1)
	if p.Min != 1 || p.Max != 2 {
		t.Errorf("the first share is %d..%d, want 1..2", p.Min, p.Max)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if a, b := gap2B12Quest(g, overseer), gap2B12Quest(g, quest); a != 1 || b != 0 {
		t.Errorf("quest counters left %d and %d, want 1 and 0", a, b)
	}
	for _, id := range []uuid.UUID{overseer, bear} {
		c, _ := g.LookupCardForEffect(id)
		if c.Counters[game.CounterPlusOne] != 1 || !game.HasKeyword(&c, "vigilance") {
			t.Errorf("%s: %d +1/+1 counters, vigilance %v", c.Name, c.Counters[game.CounterPlusOne], game.HasKeyword(&c, "vigilance"))
		}
	}
	if c, _ := g.LookupCardForEffect(theirs); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestGap2B12OverseerNotOfferedWithFewerThanThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Overseer of Vault 76", TypeLine: "Legendary Creature — Human Advisor",
		OracleID: gap2B12Overseer, Power: 3, Toughness: 3, Counters: map[string]int{"quest": 2}})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Fatalf("offered with two quest counters: %+v", c)
		}
	}
}

func TestGap2B12OverseerDeclineKeepsTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	overseer := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Overseer of Vault 76", TypeLine: "Legendary Creature — Human Advisor",
		OracleID: gap2B12Overseer, Power: 3, Toughness: 3, Counters: map[string]int{"quest": 4}})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gap2B12Quest(g, overseer); n != 4 {
		t.Errorf("%d quest counters after declining, want 4", n)
	}
}
