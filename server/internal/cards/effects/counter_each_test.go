package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_each_test.go — #1841, CR 603.2c: "whenever a [kind] counter
// is put on ~" triggers once PER COUNTER, each its own stack object
// and its own "you may".

const (
	oracleFathomMage         = "93d0e129-e3b5-4aff-9e50-f34771ed00ff"
	oracleBloodcrazedHoplite = "b1f441ce-7619-4ab5-ab36-724ed0a76728"
	oracleFlourishingDef     = "f1b26eb1-6337-42d6-a475-53ad6427f00f"
	oraclePerCounterEntrant  = "00001841-0000-4000-8000-000000000001"
)

func init() {
	// A test-only card that ENTERS with two +1/+1 counters and has
	// Fathom Mage's ability, for the CR 122.6 entry case: no real
	// Fathom Mage enters with counters on its own.
	Register(Spec{
		OracleID: oraclePerCounterEntrant,
		Name:     "Per-Counter Entrant (test)",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 2)
				return nil
			},
			Label: "test: enters with two +1/+1 counters",
		}},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverACounterIsPutOnThis(game.CounterPlusOne, "test — you may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}), "test — draw a card?"),
		},
	})
}

func pushFathomMage(g *game.Game, me uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fathom Mage", OracleID: oracleFathomMage,
		TypeLine: "Creature — Human Wizard", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordEvolve}, Owner: me, Controller: me,
	})
}

func putCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	if err := g.AddCounter(id, kind, n); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
}

// triggerPrompts is how many "you may" prompts are waiting for `who`.
func triggerPrompts(g *game.Game, who uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == who {
			n++
		}
	}
	return n
}

// answerAllPrompts says yes to every waiting prompt, and returns how
// many there were.
func answerAllPrompts(t *testing.T, g *game.Game, who uuid.UUID) int {
	t.Helper()
	n := triggerPrompts(g, who)
	for i := 0; i < n; i++ {
		answerLatestTriggerPrompt(t, g, who, true)
	}
	return n
}

func ceHandSize(g *game.Game, id uuid.UUID) int { return playerByIDForTest(g, id).Hand.Size() }

func ceSeedLibrary(g *game.Game, id uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		pushLibraryCardForTest(playerByIDForTest(g, id), game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: id})
	}
}

func TestFathomMageOneCounterIsOneTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	before := ceHandSize(g, me)
	putCounters(t, g, mage, game.CounterPlusOne, 1)
	if n := answerAllPrompts(t, g, me); n != 1 {
		t.Fatalf("%d prompts for one counter, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := ceHandSize(g, me) - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestFathomMageThreeCountersAreThreeTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	before := ceHandSize(g, me)
	putCounters(t, g, mage, game.CounterPlusOne, 3)
	if n := triggerPrompts(g, me); n != 3 {
		t.Fatalf("%d prompts for three counters, want 3 (one 'you may' each)", n)
	}
	// Decline one, accept two: each prompt is its own choice.
	answerLatestTriggerPrompt(t, g, me, false)
	answerLatestTriggerPrompt(t, g, me, true)
	answerLatestTriggerPrompt(t, g, me, true)
	passPriorityAroundTable(t, g)
	if got := ceHandSize(g, me) - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestFathomMageTriggersStackAsSeparateObjects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	putCounters(t, g, mage, game.CounterPlusOne, 2)
	answerAllPrompts(t, g, me)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me)
	mage := pushFathomMage(g, me)
	before := ceHandSize(g, me)
	// Evolve puts one counter; Doubling Season makes it two.
	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	settleProwess(t, g)
	if n := plusOnes(g, mage); n != 2 {
		t.Fatalf("%d counters under Doubling Season, want 2", n)
	}
	if n := answerAllPrompts(t, g, me); n != 2 {
		t.Fatalf("%d prompts under Doubling Season, want 2", n)
	}
	passPriorityAroundTable(t, g)
	if got := ceHandSize(g, me) - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestFathomMageHardenedScalesRaisesTheTriggerCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me)
	mage := pushFathomMage(g, me)
	// Evolve's one counter becomes two under Hardened Scales.
	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	settleProwess(t, g)
	if n := plusOnes(g, mage); n != 2 {
		t.Fatalf("%d counters under Hardened Scales, want 2", n)
	}
	if n := triggerPrompts(g, me); n != 2 {
		t.Errorf("%d prompts under Hardened Scales, want 2", n)
	}
}

func TestFathomMageIgnoresOtherKindsAndRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	mage := pushFathomMage(g, me)
	putCounters(t, g, mage, game.CounterMinusOne, 1)
	putCounters(t, g, mage, "charge", 3)
	if n := triggerPrompts(g, me); n != 0 {
		t.Fatalf("%d prompts for other kinds, want 0", n)
	}
	putCounters(t, g, mage, game.CounterPlusOne, 2)
	answerAllPrompts(t, g, me)
	passPriorityAroundTable(t, g)
	// A removal is not a placement.
	putCounters(t, g, mage, game.CounterPlusOne, -1)
	if n := triggerPrompts(g, me); n != 0 {
		t.Errorf("%d prompts for a removal, want 0", n)
	}
	// The next placement counts from the lowered total, not the old one.
	putCounters(t, g, mage, game.CounterPlusOne, 1)
	if n := triggerPrompts(g, me); n != 1 {
		t.Errorf("%d prompts for one counter after a removal, want 1", n)
	}
}

func TestFathomMageDoesNotTriggerOnOthersCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	pushFathomMage(g, me)
	bear := pushVanillaCreature(g, me, "Bear", 2, 2)
	putCounters(t, g, bear, game.CounterPlusOne, 2)
	if n := triggerPrompts(g, me); n != 0 {
		t.Errorf("%d prompts for a counter on another creature, want 0", n)
	}
}

// CR 122.6: counters a permanent enters with are put on it, and its own
// ability sees each of them.
func TestPerCounterTriggerSeesEntryCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	ceSeedLibrary(g, me, 10)
	enterCard(t, g, me, game.Card{
		Name: "Per-Counter Entrant (test)", OracleID: oraclePerCounterEntrant,
		TypeLine: "Creature — Hydra", Power: 0, Toughness: 0,
	})
	if n := triggerPrompts(g, me); n != 2 {
		t.Errorf("%d prompts for two entry counters, want 2", n)
	}
}

func TestBloodcrazedHopliteRemovesOneCounterPerCounterPut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat].ID, g.Seats[1].ID
	hoplite := pushCatalogPermanent(g, me, "Bloodcrazed Hoplite", "Creature — Human Soldier", oracleBloodcrazedHoplite, false)
	victim := pushVanillaCreature(g, opp, "Victim", 4, 4)
	putCounters(t, g, victim, game.CounterPlusOne, 5)
	putCounters(t, g, hoplite, game.CounterPlusOne, 3)
	// Three counters, three targeted triggers; each asks for its target.
	for i := 0; i < 3; i++ {
		ask := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me)
		if ask == nil {
			t.Fatalf("trigger %d did not ask for a target", i+1)
		}
		if err := g.ResolvePickTarget(ask.ID, me, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatalf("ResolvePickTarget %d: %v", i+1, err)
		}
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, victim); n != 2 {
		t.Errorf("victim has %d +1/+1 counters, want 5 - 3 = 2", n)
	}
}

func TestFlourishingDefensesMakesAnElfPerMinusCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat].ID, g.Seats[1].ID
	pushCatalogPermanent(g, me, "Flourishing Defenses", "Enchantment", oracleFlourishingDef, false)
	victim := pushVanillaCreature(g, opp, "Victim", 6, 6)
	putCounters(t, g, victim, game.CounterMinusOne, 3)
	if n := answerAllPrompts(t, g, me); n != 3 {
		t.Fatalf("%d prompts for three -1/-1 counters, want 3", n)
	}
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Elf Warrior"); n != 3 {
		t.Errorf("%d Elf Warriors, want 3", n)
	}
	// A +1/+1 counter, or a counter on a non-creature, is not it.
	putCounters(t, g, victim, game.CounterPlusOne, 1)
	if n := triggerPrompts(g, me); n != 0 {
		t.Errorf("%d prompts for a +1/+1 counter, want 0", n)
	}
}

func TestFathomMageCardRegistered(t *testing.T) {
	spec, ok := Lookup(oracleFathomMage)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].PerCounter != game.CounterPlusOne || spec.Triggered[0].OncePerBatch {
		t.Fatalf("Fathom Mage is not a per-counter, not-once-per-batch trigger: %+v", spec.Triggered)
	}
}
