package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// atq_b1_cards_test.go — Antiquities, batch 1 (#5640). Each test plays the
// real catalog entry and asserts what its text says.

const (
	atqB1WallOfSpearsOracle   = "ed836d84-ff1e-4af8-b4b8-314569b3faec"
	atqB1YotianSoldierOracle  = "645571f1-a775-4834-8b79-8852d38c9587"
	atqB1ArtifactBlastOracle  = "2d4aedc5-31c5-4281-98e1-b0c2233c3c8a"
	atqB1ReconstructionOracle = "ad8fb78b-5ca5-4ef1-8c68-ee57d1e32fec"
	atqB1ArchaeologistOracle  = "4b889ec0-6130-4e31-bb02-03fdabd28bee"
	atqB1MillstoneOracle      = "3212e47a-5492-4c50-9d4a-6ea562f1a6e1"
	atqB1JalumTomeOracle      = "32d9fd98-142b-41d2-b8f0-40ae1d4cb991"
	atqB1SageOracle           = "f34be3cc-ff47-4415-a6a8-ed142891dc0c"
	atqB1DragonEngineOracle   = "727a6474-d5b5-42ab-ace3-352d28f499eb"
	atqB1OnuletOracle         = "598f948b-bdd3-490f-b3e9-f0f9dc470522"
	atqB1AtogOracle           = "3095f2cf-05c3-4381-929f-c526b6fc30a7"
	atqB1MechanicsOracle      = "0ec58835-de2d-4064-89a9-f92db80bc276"
)

func TestAtqB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		atqB1WallOfSpearsOracle:   "Wall of Spears",
		atqB1YotianSoldierOracle:  "Yotian Soldier",
		atqB1ArtifactBlastOracle:  "Artifact Blast",
		atqB1ReconstructionOracle: "Reconstruction",
		atqB1ArchaeologistOracle:  "Argivian Archaeologist",
		atqB1MillstoneOracle:      "Millstone",
		atqB1JalumTomeOracle:      "Jalum Tome",
		atqB1SageOracle:           "Sage of Lat-Nam",
		atqB1DragonEngineOracle:   "Dragon Engine",
		atqB1OnuletOracle:         "Onulet",
		atqB1AtogOracle:           "Atog",
		atqB1MechanicsOracle:      "Orcish Mechanics",
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

// --- Wall of Spears, Yotian Soldier ----------------------------------------------

func TestAtqB1KeywordCreaturesHaveTheirPrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wall := b12Push(g, me.ID, "Wall of Spears", "Artifact Creature — Wall", atqB1WallOfSpearsOracle, 2, 3)
	soldier := b12Push(g, me.ID, "Yotian Soldier", "Artifact Creature — Soldier", atqB1YotianSoldierOracle, 1, 4)
	for _, kw := range []string{"defender", "first strike"} {
		if !hasKeywordOf(g, wall, kw) {
			t.Errorf("Wall of Spears has %s", kw)
		}
	}
	if !hasKeywordOf(g, soldier, "vigilance") {
		t.Error("Yotian Soldier has vigilance")
	}
}

// --- Artifact Blast --------------------------------------------------------------

func TestAtqB1ArtifactBlastCountersAnArtifactSpellButNotAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rock := avB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	shock := avB1Hand(me, game.Card{Name: "Shock", TypeLine: "Instant", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	refused := avB1Hand(me, game.Card{Name: "Artifact Blast", TypeLine: "Instant", OracleID: atqB1ArtifactBlastOracle})
	if err := g.CastSpell(me.ID, refused, game.CastSpellParams{Targets: cardRefs(shock)}); err == nil {
		t.Fatal("an instant spell is not an artifact spell")
	}
	blast := avB1Hand(me, game.Card{Name: "Artifact Blast", TypeLine: "Instant", OracleID: atqB1ArtifactBlastOracle})
	if err := g.CastSpell(me.ID, blast, game.CastSpellParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("Artifact Blast on the artifact spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the artifact spell is countered into its owner's graveyard")
	}
}

// --- Reconstruction, Argivian Archaeologist ---------------------------------------

func TestAtqB1ReconstructionReturnsAnArtifactCardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	bear := b17GraveyardCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	if err := castCatalogSpellErr(t, g, "Reconstruction", "Sorcery", atqB1ReconstructionOracle, cardRefs(bear)); err == nil {
		t.Fatal("a creature card is not an artifact card")
	}
	castCatalogSpell(t, g, "Reconstruction", "Sorcery", atqB1ReconstructionOracle, cardRefs(relic))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) || me.Graveyard.Contains(relic) {
		t.Error("the artifact card returns to its owner's hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card stays in the graveyard")
	}
}

func TestAtqB1ArgivianArchaeologistReturnsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := b12Push(g, me.ID, "Argivian Archaeologist", "Creature — Human Artificer", atqB1ArchaeologistOracle, 1, 1)
	relic := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	bear := b17GraveyardCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a creature card is not an artifact card")
	}
	b16Activate(t, g, me.ID, sage, 0, game.ActivateAbilityParams{Targets: cardRefs(relic)})
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card returns to its owner's hand")
	}
	if !b16Tapped(t, g, sage) {
		t.Error("the ability has a {T} cost")
	}
}

// --- Millstone -------------------------------------------------------------------

func TestAtqB1MillstoneMillsTheTargetPlayerTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "One", "Two", "Three")
	stone := b12Push(g, me.ID, "Millstone", "Artifact", atqB1MillstoneOracle, 0, 0)
	advanceToMain(t, g)
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	b16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if got := opp.Library.Size(); got != library-2 {
		t.Errorf("opponent's library %d → %d, want -2", library, got)
	}
	if got := opp.Graveyard.Size(); got != graveyard+2 {
		t.Errorf("opponent's graveyard %d → %d, want +2", graveyard, got)
	}
	if !b16Tapped(t, g, stone) {
		t.Error("the ability has a {T} cost")
	}
}

// --- Jalum Tome ------------------------------------------------------------------

func TestAtqB1JalumTomeDrawsThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drawn := seedLibrary(me, "Top Card")[0]
	pitch := avB1Hand(me, game.Card{Name: "Pitched", TypeLine: "Sorcery"})
	tome := b12Push(g, me.ID, "Jalum Tome", "Artifact — Book", atqB1JalumTomeOracle, 0, 0)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, tome, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(drawn) {
		t.Fatal("the Tome draws first")
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("after the draw, the controller discards a card")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitch}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d: draw one, discard one", hand, got)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is in the graveyard")
	}
}

// --- Sage of Lat-Nam -------------------------------------------------------------

func TestAtqB1SageOfLatNamSacrificesAnArtifactToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top Card")
	sage := b12Push(g, me.ID, "Sage of Lat-Nam", "Creature — Human Artificer", atqB1SageOracle, 1, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b12Permanent(g, me.ID, "Bear", "Creature — Bear")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature is not an artifact to sacrifice")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, sage, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the cost")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
	if !b16Tapped(t, g, sage) {
		t.Error("the ability has a {T} cost")
	}
}

// --- Dragon Engine ---------------------------------------------------------------

func TestAtqB1DragonEngineGetsPlusOneZeroPerActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	engine := b12Push(g, me.ID, "Dragon Engine", "Artifact Creature — Dragon Construct", atqB1DragonEngineOracle, 1, 3)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	b16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	if p, tough := effectivePTOf(t, g, engine); p != 3 || tough != 3 {
		t.Errorf("after two activations the Engine is %d/%d, want 3/3", p, tough)
	}
}

// --- Onulet ----------------------------------------------------------------------

func TestAtqB1OnuletGainsTwoLifeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	onulet := b12Push(g, me.ID, "Onulet", "Artifact Creature — Construct", atqB1OnuletOracle, 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(onulet) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Life; got != life+2 {
		t.Errorf("life %d → %d, want +2", life, got)
	}
}

// --- Atog ------------------------------------------------------------------------

func TestAtqB1AtogGrowsByTwoForEachArtifactItEats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	atog := b12Push(g, me.ID, "Atog", "Creature — Atog", atqB1AtogOracle, 1, 2)
	first := b12Permanent(g, me.ID, "Rock", "Artifact")
	second := b12Permanent(g, me.ID, "Stone", "Artifact")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, atog, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{first}})
	b16Activate(t, g, me.ID, atog, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{second}})
	if g.Battlefield.Contains(first) || g.Battlefield.Contains(second) {
		t.Error("each artifact is sacrificed as a cost")
	}
	if p, tough := effectivePTOf(t, g, atog); p != 5 || tough != 6 {
		t.Errorf("after two artifacts the Atog is %d/%d, want 5/6", p, tough)
	}
	if b16Tapped(t, g, atog) {
		t.Error("the ability has no {T} cost")
	}
}

// --- Orcish Mechanics ------------------------------------------------------------

func TestAtqB1OrcishMechanicsSacrificesAnArtifactToDealTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mechanics := b12Push(g, me.ID, "Orcish Mechanics", "Creature — Orc", atqB1MechanicsOracle, 1, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	life := opp.Life
	b16Activate(t, g, me.ID, mechanics, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the cost")
	}
	if got := opp.Life; got != life-2 {
		t.Errorf("opponent's life %d → %d, want -2", life, got)
	}
	if !b16Tapped(t, g, mechanics) {
		t.Error("the ability has a {T} cost")
	}
}
