package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b109_test.go — S82 GAP batch 109 (Commander Masters creatures):
// every test drives the real cast, trigger, layer and combat paths.

const (
	gap109ConstrictingOracle = "fd26f893-6f45-44df-aa4e-c919fc118aff"
	gap109HeliodOracle       = "4ea488b2-5ba0-4565-b928-570c8e03b926"
	gap109MyrOracle          = "dfee20a6-7737-49a8-b1cb-0a4914ac0ad3"
	gap109SentinelsOracle    = "d29d3b44-06dd-45dd-bbb2-506a6fce6a8d"
	gap109PiannaOracle       = "9cd8f7f6-0b69-466a-8f4c-d864ff682457"
	gap109RegalOracle        = "3a68915d-3e8b-4ede-bdd9-6c138ca63220"
	gap109SinewOracle        = "ffff90c3-63c4-4dee-a21d-6b2b113f4f80"
)

func TestGap109CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		gap109ConstrictingOracle: "Constricting Sliver",
		gap109HeliodOracle:       "Heliod, God of the Sun",
		gap109MyrOracle:          "Knighted Myr",
		gap109SentinelsOracle:    "Palace Sentinels",
		gap109PiannaOracle:       "Pianna, Nomad Captain",
		gap109RegalOracle:        "Regal Sliver",
		gap109SinewOracle:        "Sinew Sliver",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGap109SinewSliverPumpsEverySliverAtTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Push(g, me.ID, "Sinew Sliver", "Creature — Sliver", gap109SinewOracle, 1, 1)
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Sliver", TypeLine: "Creature — Sliver", Power: 2, Toughness: 2})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("Sinew Sliver power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 3 {
		t.Errorf("an opponent's Sliver has power %d, want 3 (no \"you control\")", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a non-Sliver has power %d, want 2", got)
	}
}

func TestGap109PalaceSentinelsMakesYouTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Palace Sentinels", "Creature — Human Soldier", gap109SentinelsOracle)
	if g.Monarch != uuid.Nil {
		t.Fatalf("the crown moved before the trigger resolved")
	}
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want %v", g.Monarch, me.ID)
	}
}

func TestGap109PiannaPumpsEveryAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pianna := b12Push(g, me.ID, "Pianna, Nomad Captain", "Legendary Creature — Human Nomad", gap109PiannaOracle, 2, 2)
	other := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	idle := pushVanillaCreature(g, me.ID, "Idle", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{pianna, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, pianna); got != 3 {
		t.Errorf("Pianna power %d, want 3", got)
	}
	if got := effectivePower(t, g, other); got != 3 {
		t.Errorf("the other attacker has power %d, want 3", got)
	}
	if got := effectivePower(t, g, idle); got != 2 {
		t.Errorf("a creature that did not attack has power %d, want 2", got)
	}
}

func TestGap109KnightedMyrAdaptsOnceAndGainsDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myr := b12Push(g, me.ID, "Knighted Myr", "Artifact Creature — Myr Knight", gap109MyrOracle, 2, 2)
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, myr, "double strike") {
		t.Fatal("double strike before any counter")
	}
	floatMana(t, g, me, "{W}{W}{W}")
	b16Activate(t, g, me.ID, myr, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, myr, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters = %d, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, myr, "double strike") {
		t.Fatal("no double strike after the counter was put on it")
	}
	// A second Adapt does nothing (it has a counter), so no new counters and no new trigger.
	floatMana(t, g, me, "{W}{W}{W}")
	b16Activate(t, g, me.ID, myr, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, myr, game.CounterPlusOne); got != 1 {
		t.Errorf("a second adapt added counters: %d", got)
	}
}

func TestGap109HeliodIsACreatureOnlyAtFiveDevotionAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heliod := b12Push(g, me.ID, "Heliod, God of the Sun", "Legendary Enchantment Creature — God", gap109HeliodOracle, 5, 6)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hc := gaLayered(t, g, heliod)
	if hc.IsCreature() {
		t.Fatal("Heliod is a creature with devotion below five")
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the other creatures you control lack vigilance")
	}
	for i := 0; i < 3; i++ { // three {W}{W} permanents: devotion 6
		pushTestPermanent(g, me.ID, game.Card{Name: "Pillar", TypeLine: "Enchantment", ManaCost: "{W}{W}"})
	}
	g.RunStateChecksForTest()
	hc = gaLayered(t, g, heliod)
	if !hc.IsCreature() {
		t.Error("Heliod is not a creature with devotion five")
	}
}

func TestGap109HeliodMakesAClericEnchantmentCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heliod := b12Push(g, me.ID, "Heliod, God of the Sun", "Legendary Enchantment Creature — God", gap109HeliodOracle, 5, 6)
	advanceToMain(t, g)
	floatMana(t, g, me, "{W}{W}{W}{W}")
	b16Activate(t, g, me.ID, heliod, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Cleric" && IsToken(c) {
			lc := gaLayered(t, g, c.InstanceID)
			if lc.Power != 2 || lc.Toughness != 1 || !lc.IsCreature() || !strings.Contains(lc.TypeLine, "Enchantment") {
				t.Errorf("Cleric token is %d/%d %q", lc.Power, lc.Toughness, lc.TypeLine)
			}
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Cleric tokens, want 1", n)
	}
}

func TestGap109RegalSliverCrownsTheFirstSliverAndPumpsAfterwards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	regal := b12Push(g, me.ID, "Regal Sliver", "Creature — Sliver", gap109RegalOracle, 3, 3)
	castAndResolveCreature(t, g, "Sliver A", "Creature — Sliver", "")
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want the controller of the entering Sliver", g.Monarch)
	}
	if got := effectivePower(t, g, regal); got != 3 {
		t.Fatalf("Regal Sliver power %d before the second Sliver", got)
	}
	castAndResolveCreature(t, g, "Sliver B", "Creature — Sliver", "")
	monSettle(t, g)
	// Already the monarch: the Slivers get +1/+1 instead (Regal Sliver's own, plus B's own).
	if got := effectivePower(t, g, regal); got < 4 {
		t.Errorf("Regal Sliver power %d, want at least 4 once the monarch Slivers were pumped", got)
	}
}

func TestGap109ConstrictingSliverExilesUntilThatSliverLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Constricting Sliver", "Creature — Sliver", gap109ConstrictingOracle, 3, 3)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceToMain(t, g)
	sliver := ecl6Cast(t, g, me, "Sliver A", "Creature — Sliver", "", 2, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the granted enters trigger")
	}
	if hasID(prompt.PickTargetCards, mine) {
		t.Error("the controller's own creature was offered")
	}
	pickCard(t, g, me.ID, victim)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opponent's creature was not exiled")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, sliver, 9); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Sliver left, %d of the exiled creature under its owner's control, want 1", n)
	}
}
