package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func pushFathomMage(g *game.Game, me uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fathom Mage", OracleID: effectstest.OracleFathomMage,
		TypeLine: "Creature — Human Wizard", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordEvolve}, Owner: me, Controller: me,
	})
}

func ceHandSize(g *game.Game, id uuid.UUID) int {
	return effectstest.PlayerByIDForTest(g, id).Hand.Size()
}

func TestFathomMageOneCounterIsOneTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	before := ceHandSize(g, me)
	effectstest.PutCounters(t, g, mage, game.CounterPlusOne, 1)
	if n := effectstest.AnswerAllPrompts(t, g, me); n != 1 {
		t.Fatalf("%d prompts for one counter, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := ceHandSize(g, me) - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestFathomMageThreeCountersAreThreeTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	before := ceHandSize(g, me)
	effectstest.PutCounters(t, g, mage, game.CounterPlusOne, 3)
	if n := effectstest.TriggerPrompts(g, me); n != 3 {
		t.Fatalf("%d prompts for three counters, want 3 (one 'you may' each)", n)
	}
	// Decline one, accept two: each prompt is its own choice.
	effectstest.AnswerLatestTriggerPrompt(t, g, me, false)
	effectstest.AnswerLatestTriggerPrompt(t, g, me, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := ceHandSize(g, me) - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestFathomMageTriggersStackAsSeparateObjects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	effectstest.PutCounters(t, g, mage, game.CounterPlusOne, 2)
	items := len(g.PendingTriggers)
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == mage {
			items++
		}
	}
	if items != 2 {
		t.Errorf("%d trigger objects for two counters, want 2", items)
	}
}

func TestFathomMageDoublingSeasonDoublesTheTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me)
	mage := pushFathomMage(g, me)
	before := ceHandSize(g, me)
	// Evolve puts one counter; Doubling Season makes it two.
	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, mage); n != 2 {
		t.Fatalf("%d counters under Doubling Season, want 2", n)
	}
	if n := effectstest.AnswerAllPrompts(t, g, me); n != 2 {
		t.Fatalf("%d prompts under Doubling Season, want 2", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := ceHandSize(g, me) - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestFathomMageHardenedScalesRaisesTheTriggerCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me)
	mage := pushFathomMage(g, me)
	// Evolve's one counter becomes two under Hardened Scales.
	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, mage); n != 2 {
		t.Fatalf("%d counters under Hardened Scales, want 2", n)
	}
	if n := effectstest.TriggerPrompts(g, me); n != 2 {
		t.Errorf("%d prompts under Hardened Scales, want 2", n)
	}
}

func TestFathomMageIgnoresOtherKindsAndRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	effectstest.PutCounters(t, g, mage, game.CounterMinusOne, 1)
	effectstest.PutCounters(t, g, mage, "charge", 3)
	if n := effectstest.TriggerPrompts(g, me); n != 0 {
		t.Fatalf("%d prompts for other kinds, want 0", n)
	}
	effectstest.PutCounters(t, g, mage, game.CounterPlusOne, 2)
	effectstest.AnswerAllPrompts(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	// A removal is not a placement.
	effectstest.PutCounters(t, g, mage, game.CounterPlusOne, -1)
	if n := effectstest.TriggerPrompts(g, me); n != 0 {
		t.Errorf("%d prompts for a removal, want 0", n)
	}
	// The next placement counts from the lowered total, not the old one.
	effectstest.PutCounters(t, g, mage, game.CounterPlusOne, 1)
	if n := effectstest.TriggerPrompts(g, me); n != 1 {
		t.Errorf("%d prompts for one counter after a removal, want 1", n)
	}
}

func TestFathomMageDoesNotTriggerOnOthersCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	pushFathomMage(g, me)
	bear := effectstest.PushVanillaCreature(g, me, "Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 2)
	if n := effectstest.TriggerPrompts(g, me); n != 0 {
		t.Errorf("%d prompts for a counter on another creature, want 0", n)
	}
}

// CR 122.6: counters a permanent enters with are put on it, and its own
// ability sees each of them.
func TestPerCounterTriggerSeesEntryCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.CeSeedLibrary(g, me, 10)
	effectstest.EnterCard(t, g, me, game.Card{
		Name: "Per-Counter Entrant (test)", OracleID: effectstest.OraclePerCounterEntrant,
		TypeLine: "Creature — Hydra", Power: 0, Toughness: 0,
	})
	if n := effectstest.TriggerPrompts(g, me); n != 2 {
		t.Errorf("%d prompts for two entry counters, want 2", n)
	}
}

func TestBloodcrazedHopliteRemovesOneCounterPerCounterPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat].ID, g.Seats[1].ID
	hoplite := effectstest.PushCatalogPermanent(g, me, "Bloodcrazed Hoplite", "Creature — Human Soldier", effectstest.OracleBloodcrazedHoplite, false)
	victim := effectstest.PushVanillaCreature(g, opp, "Victim", 4, 4)
	effectstest.PutCounters(t, g, victim, game.CounterPlusOne, 5)
	effectstest.PutCounters(t, g, hoplite, game.CounterPlusOne, 3)
	// Three counters, three targeted triggers; each asks for its target.
	for i := 0; i < 3; i++ {
		ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me)
		if ask == nil {
			t.Fatalf("trigger %d did not ask for a target", i+1)
		}
		if err := g.ResolvePickTarget(ask.ID, me, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatalf("ResolvePickTarget %d: %v", i+1, err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, victim); n != 2 {
		t.Errorf("victim has %d +1/+1 counters, want 5 - 3 = 2", n)
	}
}

func TestFlourishingDefensesMakesAnElfPerMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat].ID, g.Seats[1].ID
	effectstest.PushCatalogPermanent(g, me, "Flourishing Defenses", "Enchantment", effectstest.OracleFlourishingDef, false)
	victim := effectstest.PushVanillaCreature(g, opp, "Victim", 6, 6)
	effectstest.PutCounters(t, g, victim, game.CounterMinusOne, 3)
	if n := effectstest.AnswerAllPrompts(t, g, me); n != 3 {
		t.Fatalf("%d prompts for three -1/-1 counters, want 3", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Elf Warrior"); n != 3 {
		t.Errorf("%d Elf Warriors, want 3", n)
	}
	// A +1/+1 counter, or a counter on a non-creature, is not it.
	effectstest.PutCounters(t, g, victim, game.CounterPlusOne, 1)
	if n := effectstest.TriggerPrompts(g, me); n != 0 {
		t.Errorf("%d prompts for a +1/+1 counter, want 0", n)
	}
}

func TestFathomMageCardRegistered(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.OracleFathomMage)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].PerCounter != game.CounterPlusOne || spec.Triggered[0].OncePerBatch {
		t.Fatalf("Fathom Mage is not a per-counter, not-once-per-batch trigger: %+v", spec.Triggered)
	}
}
