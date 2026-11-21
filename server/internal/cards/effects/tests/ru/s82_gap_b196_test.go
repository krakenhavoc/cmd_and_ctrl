package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b196_test.go — S82 GAP batch 196 (#5948): white and black Warhammer
// 40,000 Commander cards. Each test plays the real catalog entry.

const (
	gapB196CelestineOracle    = "6503ed78-6499-47f6-856b-ee1003ce4a4f"
	gapB196ParagonOracle      = "12cc97ec-5d03-4434-a31b-51e77d208466"
	gapB196DevastatorOracle   = "fa25d142-e6ba-4fc9-81e2-114c32afc030"
	gapB196ScoutOracle        = "6b7fb630-79eb-4337-a698-626e585a495f"
	gapB196CavalryOracle      = "3b36df26-6a41-4714-95a8-fd02e2fc845f"
	gapB196TriumphOracle      = "6f8a4998-35c7-4e30-a9e7-7d4ec933238a"
	gapB196HonourGuardOracle  = "74cd099c-8476-4f75-bc1a-8b166f696163"
	gapB196VexilusOracle      = "e3a65df1-ba3a-4635-a321-8cab9a1f0740"
	gapB196ZephyrimOracle     = "d29551ab-e314-4872-9fc5-fe9a442dc8a9"
	gapB196BiotransOracle     = "5443557e-440f-4fdd-a29b-7740d7cec085"
	gapB196NurglesRotOracle   = "1b01dcff-0639-4ed0-bd9e-f28429cb803d"
	gapB196OutOfTheTombsOracl = "473a7998-669d-4181-83fe-f178a2e781f0"
)

func TestGapB196CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB196CelestineOracle:    "Celestine, the Living Saint",
		gapB196ParagonOracle:      "Grey Knight Paragon",
		gapB196DevastatorOracle:   "Space Marine Devastator",
		gapB196ScoutOracle:        "Space Marine Scout",
		gapB196CavalryOracle:      "Thunderwolf Cavalry",
		gapB196TriumphOracle:      "Triumph of Saint Katherine",
		gapB196HonourGuardOracle:  "Ultramarines Honour Guard",
		gapB196VexilusOracle:      "Vexilus Praetor",
		gapB196ZephyrimOracle:     "Zephyrim",
		gapB196BiotransOracle:     "Biotransference",
		gapB196NurglesRotOracle:   "Nurgle's Rot",
		gapB196OutOfTheTombsOracl: "Out of the Tombs",
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

func gapB196Card(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, p, tough int) uuid.UUID {
	t.Helper()
	id, err := effectstest.SpmB4Cast(t, g, name, typeLine, oracle, cost, p, tough, 0)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// --- Celestine, the Living Saint ---

func gapB196GraveCreature(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func TestGapB196CelestineReturnsACreatureCardUpToTheLifeYouGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	celestine := effectstest.B12Push(g, me.ID, "Celestine, the Living Saint", "Legendary Creature — Human Warrior", gapB196CelestineOracle, 3, 4)
	three := gapB196GraveCreature(me, "Three Drop", "{2}{G}")
	four := gapB196GraveCreature(me, "Four Drop", "{3}{G}")
	effectstest.GainLifeNow(t, g, celestine, me, 3)
	effectstest.AdvanceToEndStep(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the beginning of the end step")
	}
	if effectstest.HasID(pick.PickTargetCards, four) || !effectstest.HasID(pick.PickTargetCards, three) {
		t.Errorf("offered %v; want only the mana value 3 creature card", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the creature card did not return to the battlefield")
	}
	if !me.Graveyard.Contains(four) {
		t.Error("the mana value 4 card left the graveyard")
	}
}

func TestGapB196CelestineWithNoLifeGainedOnlyOffersManaValueZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Celestine, the Living Saint", "Legendary Creature — Human Warrior", gapB196CelestineOracle, 3, 4)
	one := gapB196GraveCreature(me, "One Drop", "{G}")
	zero := gapB196GraveCreature(me, "Zero Drop", "")
	effectstest.AdvanceToEndStep(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, one) {
			t.Error("a mana value 1 card was offered with no life gained")
		}
		effectstest.PickCard(t, g, me.ID, zero)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(one) {
		t.Error("the mana value 1 card returned")
	}
}

func TestGapB196CelestineIgnoresAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	celestine := effectstest.B12Push(g, me.ID, "Celestine, the Living Saint", "Legendary Creature — Human Warrior", gapB196CelestineOracle, 3, 4)
	theirs := gapB196GraveCreature(opp, "Their Dead", "{G}")
	effectstest.GainLifeNow(t, g, celestine, me, 5)
	effectstest.AdvanceToEndStep(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's graveyard card was offered")
	}
}

// --- Grey Knight Paragon ---

func gapB196AttackThenFlash(t *testing.T, g *game.Game, attacker uuid.UUID, name, typeLine, oracle, cost string, p, tough int) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Power: p, Toughness: tough,
		Keywords: []string{"flash"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash cast during the declare attackers step: %v", err)
	}
}

func TestGapB196GreyKnightParagonDestroysAnAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	gapB196AttackThenFlash(t, g, attacker, "Grey Knight Paragon", "Creature — Astartes Knight", gapB196ParagonOracle, "{4}{W}", 4, 4)
	for i := 0; i < 16 && effectstest.LatestPickTarget(g, me.ID) == nil && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if effectstest.HasID(pick.PickTargetCards, idle) {
			t.Error("a creature that is not attacking was offered")
		}
		effectstest.PickCard(t, g, me.ID, attacker)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !me.Graveyard.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
}

func TestGapB196GreyKnightParagonExilesAnAttackingDemon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.B12Creature(g, me.ID, "Attacking Demon", "Creature — Demon", 3, 3)
	gapB196AttackThenFlash(t, g, demon, "Grey Knight Paragon", "Creature — Astartes Knight", gapB196ParagonOracle, "{4}{W}", 4, 4)
	for i := 0; i < 16 && effectstest.LatestPickTarget(g, me.ID) == nil && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, demon)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(demon) || me.Graveyard.Contains(demon) {
		t.Error("the attacking Demon was not exiled instead of destroyed")
	}
}

// --- Space Marine Devastator ---

func TestGapB196DevastatorDestroysAnArtifactOrEnchantmentOrNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gapB196Card(t, g, "Space Marine Devastator", "Creature — Astartes Warrior", gapB196DevastatorOracle, "{3}{W}", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, bear) || !effectstest.HasID(pick.PickTargetCards, rock) {
		t.Errorf("offered %v; want the artifact only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was destroyed")
	}
}

func TestGapB196DevastatorSquadCopiesEachDestroyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Ench B", "Enchantment")
	c := effectstest.B12Permanent(g, opp.ID, "Rock C", "Artifact")
	effectstest.ToMainForCost(t, g)
	id := effectstest.TmtB11Hand(g, me, "Space Marine Devastator", "Creature — Astartes Warrior", gapB196DevastatorOracle, "{3}{W}", 3, 3)
	effectstest.FloatMana(t, g, me, "{W}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with squad paid once: %v", err)
	}
	effectstest.SettleOrderingTriggers(t, g)
	taken := map[uuid.UUID]bool{}
	for i := 0; i < 48 && !(effectstest.StackFullyEmpty(g)); i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			// Each trigger chooses its own target as it goes on the stack:
			// take the first permanent not already taken by the other one.
			var pick uuid.UUID
			for _, c := range p.PickTargetCards {
				if !taken[c] {
					pick = c
					break
				}
			}
			if pick == uuid.Nil {
				t.Fatal("no untaken permanent offered")
			}
			taken[pick] = true
			effectstest.PickCard(t, g, me.ID, pick)
			continue
		}
		if p := effectstest.TriggerOrderPromptFor(g, me.ID); p != nil {
			effectstest.SettleOrderingTriggers(t, g)
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	gone := 0
	for _, x := range []uuid.UUID{a, b, c} {
		if !g.Battlefield.Contains(x) {
			gone++
		}
	}
	if got := len(effectstest.TmtB1Named(g, "Space Marine Devastator")); got != 2 {
		t.Errorf("%d Devastators, want the card and one token copy", got)
	}
	if gone != 2 {
		t.Errorf("%d permanents destroyed, want one per Devastator (2)", gone)
	}
}

// --- Space Marine Scout ---

func gapB196Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Land", "Basic Land — Forest")
	}
}

func TestGapB196ScoutFetchesATappedPlainsWhenAnOpponentHasMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB196Lands(g, opp.ID, 3)
	gapB196Lands(g, me.ID, 1)
	plains := effectstest.SeedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	effectstest.SeedLibraryTop(me, "Other Card", "Sorcery", "{1}")
	gapB196Card(t, g, "Space Marine Scout", "Creature — Astartes Scout", gapB196ScoutOracle, "{2}{W}", 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, plains)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(plains) {
		t.Fatal("the Plains did not enter the battlefield")
	}
	if !effectstest.B16Tapped(t, g, plains) {
		t.Error("the Plains entered untapped")
	}
}

func TestGapB196ScoutDoesNothingWhenNoOpponentHasMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB196Lands(g, opp.ID, 2)
	gapB196Lands(g, me.ID, 2)
	plains := effectstest.SeedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	gapB196Card(t, g, "Space Marine Scout", "Creature — Astartes Scout", gapB196ScoutOracle, "{2}{W}", 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("a search was offered with no opponent ahead on lands")
	}
	if g.Battlefield.Contains(plains) {
		t.Error("the Plains entered anyway")
	}
}

func TestGapB196ScoutSearchIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB196Lands(g, opp.ID, 3)
	plains := effectstest.SeedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	gapB196Card(t, g, "Space Marine Scout", "Creature — Astartes Scout", gapB196ScoutOracle, "{2}{W}", 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plains) {
		t.Error("the Plains entered after the search was declined")
	}
}

// --- Thunderwolf Cavalry ---

func TestGapB196ThunderwolfCavalryPutsACounterOnEachOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cavalry := effectstest.B12Push(g, me.ID, "Thunderwolf Cavalry", "Creature — Astartes Warrior", gapB196CavalryOracle, 4, 4)
	a := effectstest.B12Creature(g, me.ID, "Friend A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Friend B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, cavalry)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("a friendly creature has %d counters, want 1", got)
		}
	}
	if got := effectstest.PlusOnes(g, cavalry); got != 0 {
		t.Errorf("the Cavalry has %d counters, want 0 (each OTHER creature)", got)
	}
	if got := effectstest.PlusOnes(g, theirs); got != 0 {
		t.Errorf("an opponent's creature has %d counters", got)
	}
}

// --- Triumph of Saint Katherine ---

func TestGapB196TriumphShufflesItselfAndTheTopSixBackOnTopOfTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.EmptyLibrary(me)
	var below, top []uuid.UUID
	for i := 0; i < 4; i++ {
		below = append(below, effectstest.SeedLibraryTop(me, "Below", "Sorcery", "{1}"))
	}
	for i := 0; i < 6; i++ {
		top = append(top, effectstest.SeedLibraryTop(me, "Top", "Sorcery", "{1}"))
	}
	triumph := effectstest.B12Push(g, me.ID, "Triumph of Saint Katherine", "Creature — Human Warrior", gapB196TriumphOracle, 5, 5)
	effectstest.KillCreature(t, g, me.ID, triumph)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(triumph) || me.Graveyard.Contains(triumph) {
		t.Fatal("the Triumph stayed in exile or the graveyard")
	}
	cards := me.Library.Cards
	if len(cards) != 11 {
		t.Fatalf("library %d cards, want 11 (ten plus the Triumph)", len(cards))
	}
	want := map[uuid.UUID]bool{triumph: true}
	for _, id := range top {
		want[id] = true
	}
	for _, c := range cards[len(cards)-7:] {
		if !want[c.InstanceID] {
			t.Errorf("the top seven hold a card (%s) that was not in the pile", c.Name)
		}
	}
	for i, id := range below {
		if cards[i].InstanceID != id {
			t.Errorf("a card below the pile moved (position %d)", i)
		}
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("%d cards left in exile", len(g.Exile.Cards))
	}
}

func TestGapB196TriumphWithAShortLibraryPutsEverythingBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.EmptyLibrary(me)
	effectstest.SeedLibraryTop(me, "Lonely", "Sorcery", "{1}")
	triumph := effectstest.B12Push(g, me.ID, "Triumph of Saint Katherine", "Creature — Human Warrior", gapB196TriumphOracle, 5, 5)
	effectstest.KillCreature(t, g, me.ID, triumph)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 2 || len(g.Exile.Cards) != 0 {
		t.Errorf("library %d, exile %d; want the Triumph and the one card back in the library", me.Library.Size(), len(g.Exile.Cards))
	}
}

func TestGapB196TriumphHasMiracleAndLifelink(t *testing.T) {
	spec, _ := effects.Lookup(gapB196TriumphOracle)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != game.AltCostKeyMiracle || spec.AlternativeCosts[0].ManaCost != "{1}{W}" {
		t.Errorf("alternative costs %+v, want miracle {1}{W}", spec.AlternativeCosts)
	}
	zeph, _ := effects.Lookup(gapB196ZephyrimOracle)
	if len(zeph.AlternativeCosts) != 1 || zeph.AlternativeCosts[0].Key != game.AltCostKeyMiracle || zeph.AlternativeCosts[0].ManaCost != "{1}{W}" {
		t.Errorf("Zephyrim alternative costs %+v, want miracle {1}{W}", zeph.AlternativeCosts)
	}
}

// --- Ultramarines Honour Guard ---

func TestGapB196HonourGuardGivesOtherCreaturesYouControlPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Ultramarines Honour Guard", "Creature — Astartes Warrior", gapB196HonourGuardOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my other creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, guard); p != 2 || tough != 2 {
		t.Errorf("the Honour Guard is %d/%d, want 2/2 (other creatures)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestGapB196HonourGuardSquadCopiesPumpEachOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.TmtB11Hand(g, me, "Ultramarines Honour Guard", "Creature — Astartes Warrior", gapB196HonourGuardOracle, "{3}{W}", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with squad paid once: %v", err)
	}
	effectstest.SettleOrderingTriggers(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.TmtB1Named(g, "Ultramarines Honour Guard")
	if len(ids) != 2 {
		t.Fatalf("%d Honour Guards, want 2", len(ids))
	}
	for _, x := range ids {
		if p, tough := effectstest.EffectivePTOf(t, g, x); p != 3 || tough != 3 {
			t.Errorf("an Honour Guard is %d/%d, want 3/3 from the other one", p, tough)
		}
	}
}

// --- Vexilus Praetor ---

func TestGapB196VexilusPraetorGivesCommandersProtectionFromEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Vexilus Praetor", "Creature — Custodes Warrior", gapB196VexilusOracle, 3, 4)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Human", 3, 3)
	effectstest.ToMainForCost(t, g)
	bolt := func(target uuid.UUID) error {
		id := effectstest.HandCardFull(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle, []string{"R"})
		return g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: false, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	}
	if err := bolt(commander); err == nil {
		t.Error("my commander was a legal target for an opponent's Bolt")
	}
	if err := bolt(bear); err != nil {
		t.Errorf("a non-commander creature was untargetable: %v", err)
	}
	if err := bolt(theirs); err != nil {
		t.Errorf("an opponent's commander was untargetable: %v", err)
	}
}

// --- Biotransference ---

func TestGapB196BiotransferenceMakesYourCreaturesArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Biotransference", "Enchantment", gapB196BiotransOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Forest")
	if c := effectstest.BattlefieldCardCopy(t, g, mine); !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("my creature is %q, want an artifact creature", c.TypeLine)
	}
	if c := effectstest.BattlefieldCardCopy(t, g, theirs); c.IsArtifact() {
		t.Error("an opponent's creature became an artifact")
	}
	if c := effectstest.BattlefieldCardCopy(t, g, land); c.IsArtifact() {
		t.Error("a land became an artifact")
	}
}

func TestGapB196BiotransferenceCreatureSpellTriggersTheDrain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Biotransference", "Enchantment", gapB196BiotransOracle, false)
	life := me.Life
	effectstest.CastWithCost(t, g, "A Creature", "Creature — Bear", "{1}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want to lose 1 for the artifact creature spell", life, me.Life)
	}
	if got := effectstest.CardsNamed(g, "Necron Warrior", me.ID); got != 1 {
		t.Fatalf("%d Necron Warrior tokens, want 1", got)
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Necron Warrior" {
			token = c
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, token.InstanceID); p != 2 || tough != 2 || !token.IsArtifact() || !token.HasColor("B") {
		t.Errorf("the token is %d/%d %q colours %v, want a 2/2 black artifact creature", p, tough, token.TypeLine, token.Colors)
	}
}

func TestGapB196BiotransferenceNoncreatureNonArtifactSpellDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Biotransference", "Enchantment", gapB196BiotransOracle, false)
	life := me.Life
	effectstest.CastWithCost(t, g, "A Sorcery", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || effectstest.CardsNamed(g, "Necron Warrior", me.ID) != 0 {
		t.Errorf("a sorcery drained me or made a token (life %d -> %d)", life, me.Life)
	}
}

func TestGapB196BiotransferenceCreatureCardsInTheGraveyardAreArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Biotransference", "Enchantment", gapB196BiotransOracle, false)
	dead := gapB196GraveCreature(me, "Dead Bear", "{1}{G}")
	got := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(dead); ok {
			got = effects.HasFold(c.Effective().Types, "Artifact")
		}
	})
	if !got {
		t.Error("a creature card in my graveyard is not an artifact")
	}
}

// --- Nurgle's Rot ---

func gapB196EnchantTheirCreature(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastCatalogSpell(t, g, "Nurgle's Rot", "Enchantment — Aura", gapB196NurglesRotOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestGapB196NurglesRotReturnsToHandAndMakesAPlaguebearerWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := gapB196EnchantTheirCreature(t, g, victim)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not enter the battlefield")
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("Nurgle's Rot did not return to its owner's hand")
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plaguebearer of Nurgle" && c.Controller == me.ID {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Plaguebearer of Nurgle token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, token.InstanceID); p != 1 || tough != 3 || !token.HasColor("B") || !token.HasSubtype("Demon") {
		t.Errorf("the token is %d/%d %q colours %v, want a 1/3 black Demon", p, tough, token.TypeLine, token.Colors)
	}
}

func TestGapB196NurglesRotCannotEnchantYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Nurgle's Rot", "Enchantment — Aura", "{B}", gapB196NurglesRotOracle, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("Nurgle's Rot was cast on my own creature")
	}
}

// --- Out of the Tombs ---

func TestGapB196OutOfTheTombsAddsTwoCountersAndMillsThatManyEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EmptyLibrary(me)
	for i := 0; i < 12; i++ {
		effectstest.SeedLibraryTop(me, "Card", "Sorcery", "{1}")
	}
	tombs := effectstest.PushCatalogPermanent(g, me.ID, "Out of the Tombs", "Enchantment", gapB196OutOfTheTombsOracl, false)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, tombs, "eon"); got != 2 {
		t.Fatalf("%d eon counters, want 2", got)
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("%d cards milled, want 2", me.Graveyard.Size())
	}
}

func TestGapB196OutOfTheTombsReturnsACreatureInsteadOfDrawingFromAnEmptyLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Out of the Tombs", "Enchantment", gapB196OutOfTheTombsOracl, false)
	effectstest.EmptyLibrary(me)
	dead := gapB196GraveCreature(me, "Dead Bear", "{1}{G}")
	other := gapB196GraveCreature(me, "Other Dead", "{2}{G}")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) || !me.Graveyard.Contains(other) {
		t.Error("the chosen creature card did not return, or the other one did")
	}
	if me.Hand.Size() != hand {
		t.Errorf("a card was drawn from an empty library (hand %d -> %d)", hand, me.Hand.Size())
	}
	if me.Eliminated || g.State != game.StateActive {
		t.Error("the game ended after a replaced draw")
	}
}

func TestGapB196OutOfTheTombsLosesTheGameWithNoCreatureToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Out of the Tombs", "Enchantment", gapB196OutOfTheTombsOracl, false)
	effectstest.EmptyLibrary(me)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.CheckState(t, g)
	if !me.Eliminated {
		t.Error("the player with no creature card to return did not lose the game")
	}
}

func TestGapB196OutOfTheTombsDoesNotApplyWhileTheLibraryHasCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Out of the Tombs", "Enchantment", gapB196OutOfTheTombsOracl, false)
	effectstest.SeedLibrary(me, "A", "B")
	gapB196GraveCreature(me, "Dead Bear", "{1}{G}")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if me.Hand.Size() != hand+1 || effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("the draw was replaced while the library had cards")
	}
}
