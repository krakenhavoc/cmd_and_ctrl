package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b227_test.go — Ixalan, batch 227 (#5532): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB227Crew     = "3d1d30e1-8934-4451-9b32-c6db2167fc2e"
	oldB227Boon     = "682eb224-cab4-4829-810b-72e63a9ceb5d"
	oldB227Sailback = "6f7b9ff4-fe79-4d21-9262-47bcd2411259"
	oldB227Shapers  = "744f79da-240f-4d14-a98c-e2796645ae88"
	oldB227Paladin  = "7e7ec48c-e31d-4ddf-a128-1969951dacf9"
	oldB227Guardian = "9128b952-bd36-4a3a-9566-34d999305991"
	oldB227Mystic   = "ca514283-c58b-4108-82b2-3b6b15b467f4"
)

func TestOldB227CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB227Crew:     "Rowdy Crew",
		oldB227Boon:     "River Heralds' Boon",
		oldB227Sailback: "Snapping Sailback",
		oldB227Shapers:  "Shapers of Nature",
		oldB227Paladin:  "Bloodcrazed Paladin",
		oldB227Guardian: "Jade Guardian",
		oldB227Mystic:   "Vineshaper Mystic",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB227Card(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// oldB227PickEach answers chooser's pick-target prompts, one per ref in the
// order the clauses are asked, then settles the stack.
func oldB227PickEach(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	for _, ref := range refs {
		p := effectstest.LatestPickTarget(g, chooser)
		if p == nil {
			effectstest.PassPriorityAroundTable(t, g)
			p = effectstest.LatestPickTarget(g, chooser)
		}
		if p == nil {
			t.Fatalf("no pick-target prompt for %v", ref)
		}
		if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB227Counters(g *game.Game, id uuid.UUID) int {
	return effectstest.CountersOn(g, id, game.CounterPlusOne)
}

// --- Rowdy Crew ---

// oldB227CrewTable empties the controller's hand and library and stacks the
// library with exactly the given type lines, so the three cards drawn are the
// only ones the random discard can pick from.
func oldB227CrewTable(t *testing.T, typeLines ...string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Library.Cards = nil
	})
	for i, tl := range typeLines {
		effectstest.SeedLibraryTop(me, "Card "+string(rune('A'+i)), tl, "")
	}
	crew := effectstest.AvB16CastCreature(t, g, "Rowdy Crew", "Creature — Human Pirate", oldB227Crew, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, crew
}

func TestOldB227RowdyCrewDrawsThreeDiscardsTwoAtRandom(t *testing.T) {
	g, me, crew := oldB227CrewTable(t, "Land", "Instant", "Sorcery")
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand is %d cards, want 1 (three drawn, two discarded)", got)
	}
	if got := me.Graveyard.Size(); got != 2 {
		t.Errorf("graveyard is %d cards, want 2", got)
	}
	if got := oldB227Counters(g, crew); got != 0 {
		t.Errorf("the Crew has %d counters after discarding cards that share no type, want 0", got)
	}
}

func TestOldB227RowdyCrewGetsTwoCountersWhenTheDiscardsShareAType(t *testing.T) {
	g, me, crew := oldB227CrewTable(t, "Land", "Basic Land — Forest", "Land")
	if got := me.Graveyard.Size(); got != 2 {
		t.Fatalf("graveyard is %d cards, want 2", got)
	}
	if got := oldB227Counters(g, crew); got != 2 {
		t.Errorf("the Crew has %d counters, want 2", got)
	}
}

func TestOldB227RowdyCrewSharedTypeNeedsOnlyOneInCommon(t *testing.T) {
	g, _, crew := oldB227CrewTable(t, "Artifact Creature — Golem", "Creature — Bear", "Creature — Wolf")
	// Any two of the three share "creature".
	if got := oldB227Counters(g, crew); got != 2 {
		t.Errorf("the Crew has %d counters, want 2", got)
	}
}

// --- Snapping Sailback ---

func TestOldB227SnappingSailbackGrowsWhenItSurvivesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saur := effectstest.B12Push(g, me.ID, "Snapping Sailback", "Creature — Dinosaur", oldB227Sailback, 4, 4)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 1, 1)
	effectstest.Cmdr2B08Damage(g, src, saur, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB227Counters(g, saur); got != 1 {
		t.Errorf("the Sailback has %d counters, want 1", got)
	}
	effectstest.Cmdr2B08Damage(g, src, saur, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB227Counters(g, saur); got != 2 {
		t.Errorf("the Sailback has %d counters after a second hit, want 2", got)
	}
}

func TestOldB227SnappingSailbackGetsNothingWhenTheDamageKillsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saur := effectstest.B12Push(g, me.ID, "Snapping Sailback", "Creature — Dinosaur", oldB227Sailback, 4, 4)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 4, 4)
	effectstest.Cmdr2B08Damage(g, src, saur, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, saur); ok {
		t.Error("the Sailback survived lethal damage")
	}
}

// --- Bloodcrazed Paladin ---

func TestOldB227BloodcrazedPaladinEntersWithACounterPerCreatureThatDied(t *testing.T) {
	for _, dead := range []int{0, 1, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < dead; i++ {
			owner := me.ID
			if i%2 == 1 {
				owner = opp.ID
			}
			victim := effectstest.B12Creature(g, owner, "Victim", "Creature — Bear", 1, 1)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
		}
		g.RunStateChecksForTest()
		paladin := effectstest.AvB16CastCreature(t, g, "Bloodcrazed Paladin", "Creature — Vampire Knight", oldB227Paladin, 1, 1)
		if got := oldB227Counters(g, paladin); got != dead {
			t.Errorf("%d creatures died: the Paladin has %d counters", dead, got)
		}
	}
}

// --- Shapers of Nature ---

func TestOldB227ShapersOfNaturePutsACounterOnTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shapers := effectstest.PushCatalogPermanent(g, me.ID, "Shapers of Nature", "Creature — Merfolk Shaman", oldB227Shapers, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, shapers, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB227Card(theirs)}}); err == nil {
		t.Fatal("activated {3}{G} with three mana")
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, shapers, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB227Card(theirs)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB227Counters(g, theirs); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
}

func TestOldB227ShapersOfNatureDrawsByRemovingACounterFromAnyOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shapers := effectstest.PushCatalogPermanent(g, me.ID, "Shapers of Nature", "Creature — Merfolk Shaman", oldB227Shapers, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, shapers, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{shapers}, CounterCounts: []int{1}}); err == nil {
		t.Fatal("activated with no +1/+1 counter on a creature to remove")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	effectstest.SeedLibraryTop(me, "Top", "Sorcery", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, shapers, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{bear}, CounterCounts: []int{1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB227Counters(g, bear); got != 0 {
		t.Errorf("the Bear still has %d counters", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

// --- Jade Guardian ---

func TestOldB227JadeGuardianCountersATargetMerfolkYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merfolk := effectstest.B12Creature(g, me.ID, "My Merfolk", "Creature — Merfolk", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	guardian := effectstest.AvB16CastCreature(t, g, "Jade Guardian", "Creature — Merfolk Shaman", oldB227Guardian, 2, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{theirs, bear} {
		if err := g.ResolvePickTarget(p.ID, me.ID, oldB227Card(bad)); err == nil {
			t.Fatalf("%s was an acceptable target", bad)
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, oldB227Card(merfolk)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB227Counters(g, merfolk) != 1 || oldB227Counters(g, theirs) != 0 || oldB227Counters(g, bear) != 0 || oldB227Counters(g, guardian) != 0 {
		t.Errorf("counters: merfolk %d theirs %d bear %d guardian %d, want 1 0 0 0",
			oldB227Counters(g, merfolk), oldB227Counters(g, theirs), oldB227Counters(g, bear), oldB227Counters(g, guardian))
	}
}

func TestOldB227JadeGuardianCanTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guardian := effectstest.AvB16CastCreature(t, g, "Jade Guardian", "Creature — Merfolk Shaman", oldB227Guardian, 2, 2)
	oldB227PickEach(t, g, me.ID, oldB227Card(guardian))
	if got := oldB227Counters(g, guardian); got != 1 {
		t.Errorf("the Guardian has %d counters, want 1", got)
	}
}

// --- Vineshaper Mystic ---

func TestOldB227VineshaperMysticCountersUpToTwoMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Merfolk A", "Creature — Merfolk", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Merfolk B", "Creature — Merfolk", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "Merfolk C", "Creature — Merfolk", 1, 1)
	effectstest.AvB16CastCreature(t, g, "Vineshaper Mystic", "Creature — Merfolk Shaman", oldB227Mystic, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{oldB227Card(a), oldB227Card(b)}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB227Counters(g, a) != 1 || oldB227Counters(g, b) != 1 || oldB227Counters(g, c) != 0 {
		t.Errorf("counters a=%d b=%d c=%d, want 1 1 0", oldB227Counters(g, a), oldB227Counters(g, b), oldB227Counters(g, c))
	}
}

// --- River Heralds' Boon ---

func TestOldB227RiverHeraldsBoonCountersACreatureAndAMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	merfolk := effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	effectstest.CastCatalogSpell(t, g, "River Heralds' Boon", "Instant", oldB227Boon, []game.TargetRef{oldB227Card(bear), oldB227Card(merfolk)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB227Counters(g, bear) != 1 || oldB227Counters(g, merfolk) != 1 {
		t.Errorf("counters bear=%d merfolk=%d, want 1 1", oldB227Counters(g, bear), oldB227Counters(g, merfolk))
	}
}

func TestOldB227RiverHeraldsBoonMerfolkClauseIsOptionalAndMustBeMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "River Heralds' Boon", "Instant", oldB227Boon, []game.TargetRef{oldB227Card(bear), oldB227Card(other)}); err == nil {
		t.Error("a non-Merfolk was accepted for the Merfolk clause")
	}
	effectstest.CastCatalogSpell(t, g, "River Heralds' Boon", "Instant", oldB227Boon, []game.TargetRef{oldB227Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB227Counters(g, bear) != 1 || oldB227Counters(g, other) != 0 {
		t.Errorf("counters bear=%d other=%d, want 1 0", oldB227Counters(g, bear), oldB227Counters(g, other))
	}
}

func TestOldB227RiverHeraldsBoonOneMerfolkMayTakeBothCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	merfolk := effectstest.B12Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	effectstest.CastCatalogSpell(t, g, "River Heralds' Boon", "Instant", oldB227Boon, []game.TargetRef{oldB227Card(merfolk), oldB227Card(merfolk)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB227Counters(g, merfolk); got != 2 {
		t.Errorf("the Merfolk has %d counters, want 2 (one per clause)", got)
	}
}
