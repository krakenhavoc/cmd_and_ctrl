package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_z1_restore_test.go — #4253 part Z1: the each-player chain
// (eachPlayerInTurn, millEachPlayerInTurn), the card-side closure mills
// (dredge, Tasha's Hideous Laughter, Consuming Aberration) and the
// per-recipient damage walk (Game.DealDamageEachEachKeyedForEffect, the
// RecipientThen riders) are keyed. Each test pauses a mill, an exile or a
// damage event on a CR 616 ordering prompt from two catalog probe
// replacements, takes a STRICT restore (the boot path), answers on the
// restored table, and checks the rest of the card ran off the disk.

const (
	z1DreadSummonsOracle = "642fa01d-025e-44dd-8360-5325e5a28282"
	z1TashasOracle       = "e352f5b9-6406-4914-bc79-f24608be6bc9"
	z1FolioOracle        = "9030d508-5cfe-46ed-954b-21d970bc3b5a"
)

// --- each-player/mill-step: Folio of Fancies ---------------------------

// Folio of Fancies with a Needle on the second opponent's library: the
// first opponent mills, the Needle's mill asks for the order, and the
// restored answer finishes that mill and the last opponent's, each by
// their hand size read as the ability began.
func TestSweepZ1FolioMillChainIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.PushLibraryTopForTest(g.Seats[2], "Needle")
	effectstest.SeedOrderProbes(g, g.Seats[2].ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	hands, libs := map[int]int{}, map[int]int{}
	for i, p := range g.Seats {
		hands[i], libs[i] = p.Hand.Size(), p.Library.Size()
	}
	spec, _ := effects.Lookup(z1FolioOracle)
	g.WithWriteLock(func() {
		if err := spec.Activated[1].Effect(g, &game.StackItem{ID: uuid.New(), Controller: me.ID}); err != nil {
			t.Fatalf("Folio: %v", err)
		}
	})
	if g.Seats[1].Library.Size() != libs[1]-hands[1] || g.Seats[3].Library.Size() != libs[3] {
		t.Fatal("the first opponent did not mill, or the last milled, before the Needle's prompt")
	}

	g = strictRestoreAtOrderPrompt(t, g)
	if !g.Seats[2].Library.Contains(needle) {
		t.Fatal("the Needle left the library before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !g.Seats[2].Graveyard.Contains(needle) {
		t.Error("the Needle was not milled after the restored answer")
	}
	for _, i := range []int{1, 2, 3} {
		if got := g.Seats[i].Library.Size(); got != libs[i]-hands[i] {
			t.Errorf("seat %d library %d, want %d: the chain did not finish off the disk", i, got, libs[i]-hands[i])
		}
	}
	if g.Seats[0].Library.Size() != libs[0] {
		t.Error("the controller milled")
	}
}

// --- the chain's done: Dread Summons -----------------------------------

// Dread Summons for X=2 with a creature card on two libraries, the second
// of them a Needle: the restored answer finishes every mill and makes one
// Zombie for each creature card milled, the chain's done tail told every
// card off the disk.
func TestSweepZ1DreadSummonsIsARestorePointAndCountsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.LibTop(g.Seats[1], game.Card{Name: "Bear One", TypeLine: "Creature — Bear"})
	needle := effectstest.LibTop(g.Seats[2], game.Card{Name: "Needle", TypeLine: "Creature — Bear"})
	effectstest.SeedOrderProbes(g, g.Seats[2].ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	last := g.Seats[3].Library.Size()
	effectstest.CastXSpell(t, g, "Dread Summons", "Instant", z1DreadSummonsOracle, "{X}{B}{B}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", me.ID); n != 0 {
		t.Fatalf("%d Zombies before the Needle's prompt was answered", n)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[2].Graveyard.Contains(needle) {
		t.Fatal("the Needle was not milled after the restored answer")
	}
	if got := g.Seats[3].Library.Size(); got != last-2 {
		t.Errorf("the last library has %d cards, want %d: the chain did not finish off the disk", got, last-2)
	}
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", g.Seats[0].ID); n != 2 {
		t.Errorf("%d Zombies, want 2 (one per creature card milled, read off the chain's carry)", n)
	}
}

// --- tashas-hideous-laughter/exile-step --------------------------------

// Tasha's Hideous Laughter with a Needle on the second opponent's library:
// the first opponent's run is done, the Needle's exile asks, and the
// restored answer finishes that run and the last opponent's.
func TestSweepZ1TashasHideousLaughterIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	needle := effectstest.PushLibraryTopForTest(g.Seats[2], "Needle")
	effectstest.SeedExileOrderProbes(g, g.Seats[2].ID)
	last := g.Seats[3].Library.Size()
	effectstest.CastCatalogSpell(t, g, "Tasha's Hideous Laughter", "Sorcery", z1TashasOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Library.Size() != 0 || g.Seats[3].Library.Size() != last {
		t.Fatal("the first opponent's run did not finish, or the last one's began, before the Needle's prompt")
	}

	g = strictRestoreAtOrderPrompt(t, g)
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, needle) {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	// The libraries are all mana value 0, so each run exiles the whole library.
	for _, i := range []int{2, 3} {
		if got := g.Seats[i].Library.Size(); got != 0 {
			t.Errorf("seat %d has %d cards left: the chain did not finish off the disk", i, got)
		}
	}
}

// --- each-player/mill-until-land-step: Consuming Aberration ------------

// Consuming Aberration's rider with a Needle on the second opponent's
// library: the restored answer mills it and the land under it, then mills
// the last opponent until a land.
func TestSweepZ1ConsumingAberrationMillChainIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.LibTop(g.Seats[2], game.Card{Name: "Under Land", TypeLine: "Basic Land — Forest"})
	needle := effectstest.LibTop(g.Seats[2], game.Card{Name: "Needle", TypeLine: "Sorcery"})
	effectstest.LibTop(g.Seats[3], game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"})
	effectstest.SeedOrderProbes(g, g.Seats[2].ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	libs := map[int]int{}
	for i, p := range g.Seats {
		libs[i] = p.Library.Size()
	}
	g.WithWriteLock(func() {
		if err := effects.B14EachOpponentMillsUntilLand(g, &game.StackItem{ID: uuid.New(), Controller: me.ID}); err != nil {
			t.Fatalf("b14EachOpponentMillsUntilLand: %v", err)
		}
	})
	if g.Seats[3].Library.Size() != libs[3] {
		t.Fatal("the last opponent milled while the Needle's prompt was open")
	}

	g = strictRestoreAtOrderPrompt(t, g)
	effectstest.AnswerExileOrder(t, g)
	if !g.Seats[2].Graveyard.Contains(needle) {
		t.Fatal("the Needle was not milled after the restored answer")
	}
	if got := g.Seats[2].Library.Size(); got != libs[2]-2 {
		t.Errorf("seat 2 library %d, want %d (the Needle and the land under it)", got, libs[2]-2)
	}
	if got := g.Seats[3].Library.Size(); got != libs[3]-1 {
		t.Errorf("seat 3 library %d, want %d: the chain did not go on off the disk", got, libs[3]-1)
	}
}

// --- dredge/return-then-draw -------------------------------------------

// A dredge whose mill reaches a Needle: the mill asks for the order, and
// the restored answer finishes the mill, returns the Imp and pays the rest
// of the draw (the parked DrawDone), so the loot's discard opens.
func TestSweepZ1DredgeMillIsARestorePointAndStillReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	needle := effectstest.PushLibraryTopForTest(p, "Needle")
	effectstest.SeedOrderProbes(g, p.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	libBefore := p.Library.Size()
	effectstest.StartLoot(g, p, 1, 1)
	effectstest.AnswerOptional(t, g, p, true)
	if effectstest.ZoneHas(p.Hand, imp) || !p.Library.Contains(needle) {
		t.Fatal("the Imp returned, or the Needle was milled, before the ordering prompt was answered")
	}

	g = strictRestoreAtOrderPrompt(t, g)
	p = g.Seats[1]
	effectstest.AnswerExileOrder(t, g)
	if !p.Graveyard.Contains(needle) || p.Library.Size() != libBefore-5 {
		t.Fatalf("after the restored answer: Needle milled %v, library %d (want %d)", p.Graveyard.Contains(needle), p.Library.Size(), libBefore-5)
	}
	if !effectstest.ZoneHas(p.Hand, imp) {
		t.Error("the Imp did not return to hand: the dredge's tail did not run off the disk")
	}
	if effectstest.DiscardPromptFor(g, p) == nil {
		t.Error("the loot's discard did not open: the parked draw was not paid")
	}
}

// --- damage-steps/next and recipient/exile-if-dealt-damage-would-die ----

// Serpentine Spike at three creatures whose controller holds the damage
// probes: every step's damage asks for the order. Restored at the first
// and again at the second, the table deals 2, 3 and 4 and exiles each
// creature instead of letting it die (the rider off the disk).
func TestSweepZ1SerpentineSpikeStepsAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	a := effectstest.P1Creature(g, opp, 1, 2)
	b := effectstest.P1Creature(g, opp, 1, 3)
	c := effectstest.P1Creature(g, opp, 1, 4)
	effectstest.SeedDamageOrderProbes(g, opp)
	effectstest.CastCatalogSpell(t, g, "Serpentine Spike", "Sorcery", effectstest.P1g3SpikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: b, Slot: 1},
		{Kind: game.TargetCard, ID: c, Slot: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	if effectstest.DamageMarkedOn(g, b) != 0 {
		t.Fatal("the second step was dealt before its own prompt")
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for id, what := range map[uuid.UUID]string{a: "the 2-damage target", b: "the 3-damage target", c: "the 4-damage target"} {
		effectstest.P1WantZone(t, g, id, game.ZoneExile, what)
	}
}

// --- red-suns-zenith/shuffle-into-library ------------------------------

// Red Sun's Zenith at a bear whose controller holds the probes: the
// restored answer deals the 2, exiles the bear instead of letting it die
// (the per-recipient rider) and shuffles the Zenith into its owner's
// library (the walk's total tail).
func TestSweepZ1RedSunsZenithIsARestorePointAndStillShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.P1Creature(g, opp.ID, 2, 2)
	effectstest.SeedDamageOrderProbes(g, opp.ID)
	z := effectstest.CastXSpell(t, g, "Red Sun's Zenith", "Sorcery", effectstest.P1g2ZenithOracle, "{X}{R}", 2, effectstest.Pr6Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, bear) != 0 {
		t.Fatal("the damage was dealt before the ordering prompt was answered")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear the Zenith killed")
	effectstest.P1WantZone(t, g, z, game.ZoneLibrary, "Red Sun's Zenith")
}
