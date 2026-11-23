package testsg

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap_b173_cards_test.go — Jumpstart 2022 cards, S82 GAP batch 173
// (#5948). Every test plays the real catalog entry through the engine.

const (
	gapNoEscapeOracle             = "7c6e0198-edd8-42b8-ba1b-549e1713d188"
	gapPerilousVoyageOracle       = "0ab090db-f8fc-42ff-95db-e9370796db97"
	gapStartlingDevelopmentOracle = "8747ca90-ac1d-45fd-89fe-aac385cd66b4"
	gapSynchronizedEvictionOracle = "ad64b05a-4702-47ef-a81e-1b481e57ab55"
	gapTragicLessonOracle         = "516d1ac4-e831-44b6-8266-131dde7fe802"
	gapBuiltToLastOracle          = "403e554b-72a5-48b3-acea-0b4062ffb913"
	gapFlickerOfFateOracle        = "d5b9d250-92f5-40d8-b962-4ae4adb466c8"
	gapJusticiarsPortalOracle     = "0214805d-d207-42da-a25f-2a8e1990904e"
)

func TestGapB173CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapNoEscapeOracle: "No Escape", gapPerilousVoyageOracle: "Perilous Voyage",
		gapStartlingDevelopmentOracle: "Startling Development", gapSynchronizedEvictionOracle: "Synchronized Eviction",
		gapTragicLessonOracle: "Tragic Lesson", gapBuiltToLastOracle: "Built to Last",
		gapFlickerOfFateOracle: "Flicker of Fate", gapJusticiarsPortalOracle: "Justiciar's Portal",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registration %+v", name, spec)
		}
	}
}

func TestGapNoEscapeCountersAndExilesACreatureSpellAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fox := effectstest.CastCatalogSpell(t, g, "Devilthorn Fox", "Creature — Fox", "8f80a7be-3da3-4e84-82a8-dd5d4b2c3f32", nil)
	instantSpell := effectstest.CastCatalogSpell(t, g, "Plain Instant", "Instant", "bbbbbbbb-0000-0000-0000-000000000001", nil)
	refused := effectstest.PushCatalogHandCard(opp, "No Escape", "Instant", gapNoEscapeOracle)
	if err := g.CastSpell(opp.ID, refused, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(instantSpell)}); err == nil {
		t.Fatal("No Escape accepted a spell that is neither a creature nor a planeswalker")
	}
	spell := effectstest.PushCatalogHandCard(opp, "No Escape", "Instant", gapNoEscapeOracle)
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(fox)}); err != nil {
		t.Fatalf("cast No Escape: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fox) || me.Graveyard.Contains(fox) || !g.Exile.Contains(fox) {
		t.Error("the countered creature spell is not in exile")
	}
	if effectstest.ScryChoiceFor(g, opp.ID) == nil {
		t.Error("no scry 1 for the controller")
	}
}

func TestGapPerilousVoyageScriesOnlyForSmallPermanents(t *testing.T) {
	for _, tc := range []struct {
		cost string
		scry bool
	}{{"{1}{U}", true}, {"{2}{U}", false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		perm := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Thing", TypeLine: "Artifact", ManaCost: tc.cost})
		land := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Land", TypeLine: "Land"})
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		for _, bad := range []uuid.UUID{land, mine} {
			if err := effectstest.CastCatalogSpellErr(t, g, "Perilous Voyage", "Instant", gapPerilousVoyageOracle, effectstest.CardTarget(bad)); err == nil {
				t.Fatal("Perilous Voyage took a land or a permanent I control")
			}
		}
		effectstest.CastCatalogSpell(t, g, "Perilous Voyage", "Instant", gapPerilousVoyageOracle, effectstest.CardTarget(perm))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(perm) || !opp.Hand.Contains(perm) {
			t.Errorf("%s: the permanent is not in its owner's hand", tc.cost)
		}
		if got := effectstest.ScryChoiceFor(g, me.ID) != nil; got != tc.scry {
			t.Errorf("%s: scry prompt = %v, want %v", tc.cost, got, tc.scry)
		}
	}
}

func TestGapStartlingDevelopmentMakesABlueSerpentFourFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Artifact Bear", TypeLine: "Artifact Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	effectstest.CastCatalogSpell(t, g, "Startling Development", "Instant", gapStartlingDevelopmentOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 4 || tou != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tou)
	}
	c := effectstest.LayeredCard(t, g, bear)
	if !slices.Equal(c.Effective().Colors, []string{"U"}) {
		t.Errorf("colours = %v, want blue only", c.Effective().Colors)
	}
	if st := c.Effective().Subtypes; !slices.Equal(st, []string{"Serpent"}) {
		t.Errorf("subtypes = %v, want Serpent only", st)
	}
	if !c.IsArtifact() || !c.IsCreature() {
		t.Error("it lost a card type")
	}
	effectstest.EndTurn(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 2 || tou != 2 {
		t.Errorf("after the turn P/T = %d/%d, want 2/2", p, tou)
	}
}

func TestGapStartlingDevelopmentCyclesForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Startling Development", "Instant", gapStartlingDevelopmentOracle, "{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
}

func TestGapSynchronizedEvictionCostsTwoLessWithTwoSharedTypes(t *testing.T) {
	cast := func(t *testing.T, pool []game.ManaToken, setup func(g *game.Game, me *game.Player)) (*game.Game, uuid.UUID, error) {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		setup(g, me)
		target := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Thing", TypeLine: "Artifact"})
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Synchronized Eviction", TypeLine: "Instant", OracleID: gapSynchronizedEvictionOracle,
			ManaCost: "{4}{U}", Owner: me.ID, Controller: me.ID})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatal(err)
			}
		}
		me.ManaPool.AddMana(pool...)
		return g, target, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	cheap := []game.ManaToken{{Color: "U"}, {Color: "C"}, {Color: "C"}}
	full := []game.ManaToken{{Color: "U"}, {Color: "C"}, {Color: "C"}, {Color: "C"}, {Color: "C"}}
	twoBears := func(g *game.Game, me *game.Player) {
		effectstest.B12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
		effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear Druid", 2, 2)
	}
	if _, _, err := cast(t, cheap, func(g *game.Game, me *game.Player) {
		effectstest.B12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
		effectstest.B12Creature(g, me.ID, "Elf One", "Creature — Elf", 2, 2)
	}); err == nil {
		t.Error("cast for {2}{U} with no two creatures sharing a type")
	}
	g, target, err := cast(t, cheap, twoBears)
	if err != nil {
		t.Fatalf("cast for {2}{U} with two Bears: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Fatal("the target is still on the battlefield")
	}
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if idx := effectstest.LibraryIndex(owner, target); idx != len(owner.Library.Cards)-2 {
		t.Errorf("the target is at library index %d, want second from the top (%d)", idx, len(owner.Library.Cards)-2)
	}
	if _, _, err := cast(t, full, func(*game.Game, *game.Player) {}); err != nil {
		t.Errorf("full price cast refused: %v", err)
	}
}

func TestGapTragicLessonDiscardsUnlessALandReturns(t *testing.T) {
	// Declining the return is the discard.
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Land", TypeLine: "Basic Land — Forest"})
	handBefore := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Tragic Lesson", "Instant", gapTragicLessonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no land pick offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the return: %v", err)
	}
	if c := effectstest.DiscardChoiceFor(g, me.ID); c == nil {
		t.Fatal("declining the return did not ask for a discard")
	}
	if len(me.Hand.Cards) != handBefore+2 {
		t.Errorf("hand = %d, want %d (two drawn)", len(me.Hand.Cards), handBefore+2)
	}
}

func TestGapTragicLessonReturningALandSkipsTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Land", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Tragic Lesson", "Instant", gapTragicLessonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no land pick offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("returning the land: %v", err)
	}
	if g.Battlefield.Contains(land) || !me.Hand.Contains(land) {
		t.Error("the land did not return to hand")
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("asked for a discard after returning a land")
	}
}

func TestGapTragicLessonWithNoLandIsTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Tragic Lesson", "Instant", gapTragicLessonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Error("no discard with no land to return")
	}
}

func TestGapBuiltToLastPumpsAndProtectsOnlyAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	plain := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	golem := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3})
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Land", TypeLine: "Land"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Built to Last", "Instant", gapBuiltToLastOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Built to Last targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Built to Last", "Instant", gapBuiltToLastOracle, effectstest.CardTarget(plain))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Built to Last", "Instant", gapBuiltToLastOracle, effectstest.CardTarget(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, plain); p != 4 || tou != 4 {
		t.Errorf("Bear = %d/%d, want 4/4", p, tou)
	}
	if p, tou := effectstest.PtOf(t, g, golem); p != 5 || tou != 5 {
		t.Errorf("Golem = %d/%d, want 5/5", p, tou)
	}
	if effectstest.P1g2HasKeyword(g, plain, "indestructible") {
		t.Error("a non-artifact creature gained indestructible")
	}
	if !effectstest.P1g2HasKeyword(g, golem, "indestructible") {
		t.Error("the artifact creature did not gain indestructible")
	}
}

func TestGapFlickerOfFateReturnsACreatureOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Counters: map[string]int{"+1/+1": 1}})
	art := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Trinket", TypeLine: "Artifact"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Flicker of Fate", "Instant", gapFlickerOfFateOracle, effectstest.CardTarget(art)); err == nil {
		t.Fatal("Flicker of Fate targeted an artifact")
	}
	effectstest.CastCatalogSpell(t, g, "Flicker of Fate", "Instant", gapFlickerOfFateOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	back, ok := effectstest.VisBattlefieldNamed(g, opp.ID, "Their Bear")
	if !ok {
		t.Fatal("the creature did not return under its owner's control")
	}
	if effectstest.CountersOn(g, back.InstanceID, "+1/+1") != 0 {
		t.Error("it kept its counter: it is not a new object")
	}
}

func TestGapJusticiarsPortalFlickersAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Counters: map[string]int{"+1/+1": 1}})
	theirs := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := effectstest.CastCatalogSpellErr(t, g, "Justiciar's Portal", "Instant", gapJusticiarsPortalOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Justiciar's Portal targeted a creature I don't control")
	}
	effectstest.CastCatalogSpell(t, g, "Justiciar's Portal", "Instant", gapJusticiarsPortalOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	back, ok := effectstest.VisBattlefieldNamed(g, me.ID, "My Bear")
	if !ok {
		t.Fatal("the creature did not return")
	}
	if effectstest.CountersOn(g, back.InstanceID, "+1/+1") != 0 {
		t.Error("it kept its counter: it is not a new object")
	}
	if !effectstest.P1g2HasKeyword(g, back.InstanceID, "first strike") {
		t.Error("the returned creature has no first strike")
	}
	effectstest.EndTurn(t, g)
	if effectstest.P1g2HasKeyword(g, back.InstanceID, "first strike") {
		t.Error("first strike outlasted the turn")
	}
}
