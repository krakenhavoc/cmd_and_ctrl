package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_6149_protection_test.go — #6149's two protection qualities through
// the cards that print them: Geyadrone Dihada's "protection from
// permanents with corruption counters on them" and Rebbec, Architect of
// Ascension's "protection from each mana value among artifacts you
// control" (CR 702.16a, 702.16i).

const (
	geyadroneOracle = "e1105d84-86b5-44ef-9efc-d994ee1c35cc"
	rebbecOracle    = "ccad9b56-3c31-445f-8b78-b5bbdb3d077a"
	prodigalOracle  = "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe"
)

func TestS82Seam6149CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		geyadroneOracle: "Geyadrone Dihada",
		rebbecOracle:    "Rebbec, Architect of Ascension",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: spec %+v, registered %v", name, spec, ok)
		}
	}
}

func geyadroneTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Geyadrone Dihada", "Legendary Planeswalker — Dihada", geyadroneOracle, 4)
	return g, me, opp, walker
}

func addCounter(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, kind, n) })
	if err != nil {
		t.Fatalf("AddCounterForEffect(%s, %d): %v", kind, n, err)
	}
}

// A permanent with a corruption counter can't target Geyadrone, and can
// once the counter is gone.
func TestGeyadroneCantBeTargetedByACorruptedPermanent(t *testing.T) {
	g, me, _, walker := geyadroneTable(t)
	sorcerer := effectstest.B27Push(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard", prodigalOracle, "{2}{U}", 1, 1, "U")
	addCounter(t, g, sorcerer, "corruption", 1)
	if err := effectstest.InvB1Activate(g, me.ID, sorcerer, 0, effectstest.CardTarget(walker)...); err == nil {
		t.Fatal("a permanent with a corruption counter targeted Geyadrone Dihada")
	}
	addCounter(t, g, sorcerer, "corruption", -1)
	if err := effectstest.InvB1Activate(g, me.ID, sorcerer, 0, effectstest.CardTarget(walker)...); err != nil {
		t.Fatalf("the counter is gone, yet the target was refused: %v", err)
	}
}

// Damage from a corrupted permanent is prevented (CR 702.16e); from a
// clean one it is dealt.
func TestGeyadronePreventsDamageFromACorruptedPermanent(t *testing.T) {
	g, _, opp, walker := geyadroneTable(t)
	bad := effectstest.B27Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	addCounter(t, g, bad, "corruption", 2)
	effectstest.Damage(t, g, bad, walker, 2)
	g.RunStateChecksForTest()
	if got := effectstest.LoyaltyCount(g, walker); got != 4 {
		t.Errorf("loyalty after damage from a corrupted Ogre = %d, want 4", got)
	}
	clean := effectstest.B27Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	effectstest.Damage(t, g, clean, walker, 2)
	g.RunStateChecksForTest()
	if got := effectstest.LoyaltyCount(g, walker); got != 2 {
		t.Errorf("loyalty after 2 damage from a clean Ogre = %d, want 2", got)
	}
}

func TestGeyadronePlusOneDrainsAndCorrupts(t *testing.T) {
	g, me, opp, walker := geyadroneTable(t)
	bear := effectstest.B27Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	mine, theirs := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	if err := effectstest.InvB1Activate(g, me.ID, walker, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID) - mine; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if got := theirs - effectstest.LifeOf(g, opp.ID); got != 2 {
		t.Errorf("the opponent lost %d life, want 2", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, bear).Counters["corruption"]; got != 1 {
		t.Errorf("corruption counters on the Bear = %d, want 1", got)
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
	// "other": Geyadrone can't put the counter on itself.
	g2, me2, _, walker2 := geyadroneTable(t)
	if err := effectstest.InvB1Activate(g2, me2.ID, walker2, 0, effectstest.CardTarget(walker2)...); err == nil {
		t.Error("the +1 targeted Geyadrone Dihada itself")
	}
}

func TestGeyadroneMinusThreeThreatensAndCorrupts(t *testing.T) {
	g, me, opp, walker := geyadroneTable(t)
	bear := effectstest.B27Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := effectstest.InvB1Activate(g, me.ID, walker, 1, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindCardAnywhere(t, g, bear)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Error("Geyadrone's controller does not control the Bear")
	}
	if c.Tapped {
		t.Error("the Bear was not untapped")
	}
	if c.Counters["corruption"] != 1 {
		t.Errorf("corruption counters = %d, want 1", c.Counters["corruption"])
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the Bear does not have haste")
	}
	effectstest.RunCleanup(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func TestGeyadroneMinusSevenTakesEveryCorruptedPermanent(t *testing.T) {
	g, me, opp, walker := geyadroneTable(t)
	addCounter(t, g, walker, game.CounterLoyalty, 4)
	a := effectstest.B27Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	b := effectstest.B27Push(g, opp.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	clean := effectstest.B27Push(g, opp.ID, "Elk", "Creature — Elk", "", "{1}{G}", 2, 2, "G")
	addCounter(t, g, a, "corruption", 1)
	addCounter(t, g, b, "corruption", 1)
	if err := effectstest.InvB1Activate(g, me.ID, walker, 2); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.ControllerOf(t, g, id) != me.ID {
			t.Errorf("a corrupted permanent %s was not taken", id)
		}
	}
	if effectstest.ControllerOf(t, g, clean) != opp.ID {
		t.Error("a permanent with no corruption counter was taken")
	}
	effectstest.RunCleanup(t, g)
	if effectstest.ControllerOf(t, g, a) != me.ID {
		t.Error("the −7's control ended at end of turn; it has no stated duration")
	}
}

// castWithCost casts a catalog spell that carries its printed cost, so its
// mana value is known on the stack.
func castWithCost(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, target uuid.UUID) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target)})
}

// With artifacts of mana value 1 and 3 under Rebbec, a mana value 3 source
// can't target them and a 2 can (CR 702.16i: one protection per value).
func TestRebbecGivesArtifactsProtectionFromEachManaValueAmongThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B27Push(g, me.ID, "Rebbec, Architect of Ascension", "Legendary Creature — Human Artificer", rebbecOracle, "{3}{W}", 3, 4, "W")
	one := effectstest.B27Push(g, me.ID, "Myr", "Artifact Creature — Myr", "", "{1}", 1, 1)
	effectstest.B27Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", "{3}", 3, 3)
	if err := castWithCost(t, g, "Murder", "Instant", "938b4e2c-88d9-4637-bc00-e228920c9a78", "{1}{B}{B}", one); err == nil {
		t.Fatal("a mana value 3 spell targeted an artifact while a mana value 3 artifact is under Rebbec")
	}
	if err := castWithCost(t, g, "Searing Spear", "Instant", "aafd44c1-74a9-4aa7-a4a2-67eb39a07478", "{1}{R}", one); err != nil {
		t.Fatalf("a mana value 2 spell could not target the artifact: %v", err)
	}
}

// The set is read as the check is made: once the mana value 3 artifact is
// gone, a mana value 3 spell can target the one left. Nonartifacts and an
// opponent's artifacts get nothing.
func TestRebbecProtectionTracksTheBoard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rebbec := effectstest.B27Push(g, me.ID, "Rebbec, Architect of Ascension", "Legendary Creature — Human Artificer", rebbecOracle, "{3}{W}", 3, 4, "W")
	one := effectstest.B27Push(g, me.ID, "Myr", "Artifact Creature — Myr", "", "{1}", 1, 1)
	golem := effectstest.B27Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", "{3}", 3, 3)
	theirs := effectstest.B27Push(g, opp.ID, "Golem", "Artifact Creature — Golem", "", "{3}", 3, 3)
	got := effectstest.ProtectionsOn(t, g, one)
	if len(got) != 2 {
		t.Errorf("protections on the Myr = %v, want mana value 1 and mana value 3", got)
	}
	if p := effectstest.ProtectionsOn(t, g, rebbec); len(p) != 0 {
		t.Errorf("Rebbec (not an artifact) has protection %v", p)
	}
	if p := effectstest.ProtectionsOn(t, g, theirs); len(p) != 0 {
		t.Errorf("an opponent's artifact has protection %v", p)
	}
	effectstest.LeaveBattlefield(t, g, me.ID, golem)
	if err := castWithCost(t, g, "Murder", "Instant", "938b4e2c-88d9-4637-bc00-e228920c9a78", "{1}{B}{B}", one); err != nil {
		t.Fatalf("with no mana value 3 artifact left, Murder could not target the Myr: %v", err)
	}
}
