package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b189_test.go — S82 GAP batch 189 (#5948), Warhammer 40,000
// Commander creatures. Each test plays the real catalog entry.

const (
	gapB189PriestOracle        = "22f5f39d-0898-4c47-8fa8-7d97102d3229"
	gapB189ImmortalOracle      = "7388a797-d40d-4783-a281-d17bc159b856"
	gapB189NightbringerOracle  = "eedd5a17-d0fe-4a13-8f2c-f343bfa43f92"
	gapB189VoidDragonOracle    = "d04d2434-121e-4420-abfc-fa21ad5577c3"
	gapB189DestroyerOracle     = "f96fcb37-5f17-4231-b1a2-64f5fe45730b"
	gapB189LordOracle          = "c40e8d52-d4cc-4502-8b51-a5612dbe1ada"
	gapB189BilepiperOracle     = "6bd38092-7788-4bb6-9447-16c56e11e603"
	gapB189SzarekhOracle       = "9a63f950-2e8a-4617-896c-1d0af538dbc7"
	gapB189TallymanOracle      = "7cc4a635-8730-4018-a8fd-b29fb5479928"
	gapB189TechnomancerOracle  = "3f97541e-34f3-4f7f-8ac2-7167b2ec9b33"
	gapB189TombBladeOracle     = "c857de66-f4e0-4e97-ad01-fac623af8225"
	gapB189PraetorianOracle    = "52aa5a79-db0e-4eaa-b2df-e41674c47794"
	gapB189ArtifactCreatureTL  = "Artifact Creature — Necron"
	gapB189ArtifactCreatureTL2 = "Artifact Creature — Construct"
)

func TestGapB189CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB189PriestOracle:       "Sanguinary Priest",
		gapB189ImmortalOracle:     "Sautekh Immortal",
		gapB189NightbringerOracle: "Shard of the Nightbringer",
		gapB189VoidDragonOracle:   "Shard of the Void Dragon",
		gapB189DestroyerOracle:    "Skorpekh Destroyer",
		gapB189LordOracle:         "Skorpekh Lord",
		gapB189BilepiperOracle:    "Sloppity Bilepiper",
		gapB189SzarekhOracle:      "Szarekh, the Silent King",
		gapB189TallymanOracle:     "Tallyman of Nurgle",
		gapB189TechnomancerOracle: "Technomancer",
		gapB189TombBladeOracle:    "Tomb Blade",
		gapB189PraetorianOracle:   "Triarch Praetorian",
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

// --- Sanguinary Priest ---------------------------------------------------

func TestGapB189PriestPingsWhenAnotherCreatureOfYoursDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sanguinary Priest", "Creature — Astartes Cleric", gapB189PriestOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want 1 damage", life, opp.Life)
	}
}

func TestGapB189PriestIgnoresItsOwnDeathAndOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := effectstest.B12Push(g, me.ID, "Sanguinary Priest", "Creature — Astartes Cleric", gapB189PriestOracle, 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("an opponent's creature dying put a target prompt up")
		}
	}
	effectstest.KillCreature(t, g, me.ID, priest)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the Priest's own death put a target prompt up")
		}
	}
}

// --- Sautekh Immortal ----------------------------------------------------

func TestGapB189ImmortalEntersWithACounterPerCreatureThatDied(t *testing.T) {
	for _, died := range []int{0, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		for i := 0; i < died; i++ {
			c := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
			effectstest.KillCreature(t, g, opp.ID, c)
		}
		id := effectstest.AvB16CastCreature(t, g, "Sautekh Immortal", gapB189ArtifactCreatureTL, gapB189ImmortalOracle, 2, 2)
		_ = me
		if got := effectstest.PlusOnes(g, id); got != died {
			t.Errorf("%d creatures died: %d counters, want %d", died, got, died)
		}
	}
}

// --- Triarch Praetorian --------------------------------------------------

func TestGapB189PraetorianDrawsTwoLosesTwoWhenUnearthed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	effectstest.BroB04Unearth(t, g, effectstest.BroB04Card{
		Name: "Triarch Praetorian", TypeLine: gapB189ArtifactCreatureTL, Oracle: gapB189PraetorianOracle,
		Power: 2, Toughness: 1, Unearth: "{4}{B}",
	})
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want 2 cards drawn", hand, me.Hand.Size())
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 lost", life, me.Life)
	}
}

func TestGapB189PraetorianDrawsNothingWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	life, lib := me.Life, me.Library.Size()
	effectstest.AvB16CastCreature(t, g, "Triarch Praetorian", gapB189ArtifactCreatureTL, gapB189PraetorianOracle, 2, 1)
	if me.Life != life || me.Library.Size() != lib {
		t.Errorf("a cast Praetorian drew: life %d -> %d, library %d -> %d", life, me.Life, lib, me.Library.Size())
	}
}

// --- Skorpekh Destroyer --------------------------------------------------

func TestGapB189DestroyerGainsFirstStrikeWhenAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	destroyer := effectstest.B12Push(g, me.ID, "Skorpekh Destroyer", gapB189ArtifactCreatureTL, gapB189DestroyerOracle, 4, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, destroyer, "first strike") {
		t.Fatal("the Destroyer has first strike before an artifact entered")
	}
	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "Rock", TypeLine: "Artifact"}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, destroyer, "first strike") {
		t.Error("the Destroyer did not gain first strike")
	}
}

func TestGapB189DestroyerGainsItWhenItselfEntersButNotForNonartifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.AvB16CastCreature(t, g, "Skorpekh Destroyer", gapB189ArtifactCreatureTL, gapB189DestroyerOracle, 4, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "first strike") {
		t.Error("the Destroyer entering is an artifact entering under your control")
	}
	_ = me

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	d2 := effectstest.B12Push(g2, me2.ID, "Skorpekh Destroyer", gapB189ArtifactCreatureTL, gapB189DestroyerOracle, 4, 2)
	effectstest.AvB16CastCreature(t, g2, "Bear", "Creature — Bear", "", 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g2, d2, "first strike") {
		t.Error("a nonartifact creature entering gave first strike")
	}
}

// --- Skorpekh Lord -------------------------------------------------------

func TestGapB189LordPumpsOtherArtifactCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Skorpekh Lord", gapB189ArtifactCreatureTL, gapB189LordOracle, 3, 2)
	mine := effectstest.B12Creature(g, me.ID, "Bot", gapB189ArtifactCreatureTL2, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bot", gapB189ArtifactCreatureTL2, 2, 2)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("my artifact creature has power %d, want 3", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "menace") {
		t.Error("my artifact creature did not gain menace")
	}
	if p := effectstest.EffectivePower(t, g, lord); p != 3 {
		t.Errorf("the Lord pumped itself: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Errorf("a nonartifact creature was affected (power %d)", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "menace") {
		t.Errorf("an opponent's artifact creature was affected (power %d)", p)
	}
}

func TestGapB189LordUnearthReturnsItWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, _ := effectstest.BroB04Unearth(t, g, effectstest.BroB04Card{
		Name: "Skorpekh Lord", TypeLine: gapB189ArtifactCreatureTL, Oracle: gapB189LordOracle,
		Power: 3, Toughness: 2, Unearth: "{2}{B}",
	})
	if effectstest.B12ZoneOf(g, id) != game.ZoneBattlefield {
		t.Errorf("the Lord is in %v after unearth", effectstest.B12ZoneOf(g, id))
	}
}

// --- Tallyman of Nurgle --------------------------------------------------

func gapB189TallymanTable(t *testing.T, died int) (*game.Game, *game.Player, int, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tallyman of Nurgle", "Creature — Astartes Warrior", gapB189TallymanOracle, 2, 3)
	names := make([]string, 10)
	for i := range names {
		names[i] = "Card"
	}
	effectstest.SeedLibrary(me, names...)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < died; i++ {
		c := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
		effectstest.KillCreature(t, g, opp.ID, c)
	}
	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, me.Hand.Size() - hand, life - me.Life
}

func TestGapB189TallymanDrawsOneAfterADeath(t *testing.T) {
	_, _, drew, lost := gapB189TallymanTable(t, 1)
	if drew != 1 || lost != 1 {
		t.Errorf("drew %d, lost %d, want 1 and 1", drew, lost)
	}
}

func TestGapB189TallymanDoesNothingWithoutADeath(t *testing.T) {
	_, _, drew, lost := gapB189TallymanTable(t, 0)
	if drew != 0 || lost != 0 {
		t.Errorf("drew %d, lost %d, want nothing", drew, lost)
	}
}

func TestGapB189TallymanDrawsSevenWhenSevenDied(t *testing.T) {
	_, _, drew, lost := gapB189TallymanTable(t, 7)
	if drew != 7 || lost != 7 {
		t.Errorf("drew %d, lost %d, want 7 and 7 (instead of 1 and 1)", drew, lost)
	}
}

func TestGapB189TallymanSixIsStillOne(t *testing.T) {
	_, _, drew, lost := gapB189TallymanTable(t, 6)
	if drew != 1 || lost != 1 {
		t.Errorf("drew %d, lost %d, want 1 and 1 for six deaths", drew, lost)
	}
}

// --- Shard of the Nightbringer -------------------------------------------

func TestGapB189NightbringerHalvesAnOpponentWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = 21
	life := me.Life
	effectstest.AvB16CastCreature(t, g, "Shard of the Nightbringer", "Creature — C'tan", gapB189NightbringerOracle, 8, 8)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("opponent life %d, want 10 after losing half of 21 rounded up (11)", opp.Life)
	}
	if me.Life != life+11 {
		t.Errorf("my life %d -> %d, want +11", life, me.Life)
	}
}

func TestGapB189NightbringerDoesNothingWhenNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	oppLife := opp.Life
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Shard of the Nightbringer", TypeLine: "Creature — C'tan",
			OracleID: gapB189NightbringerOracle, Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatal(err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("a Nightbringer put onto the battlefield asked for a target")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Errorf("opponent life %d -> %d", oppLife, opp.Life)
	}
}

// --- Shard of the Void Dragon --------------------------------------------

func TestGapB189VoidDragonAttackMakesEachOpponentSacrificeANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Shard of the Void Dragon", "Creature — C'tan", gapB189VoidDragonOracle, 7, 7)
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWithTrigger(t, g, dragon, opp.ID, uuid.Nil)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent kept their only nonland permanent")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == opp.ID {
			return
		}
	}
	t.Error("the opponent's land was sacrificed")
}

func TestGapB189VoidDragonGrowsForArtifactsLeavingToGraveyardOrExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Shard of the Void Dragon", "Creature — C'tan", gapB189VoidDragonOracle, 7, 7)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	rock2 := effectstest.B12Permanent(g, me.ID, "Other Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, dragon); got != 2 {
		t.Fatalf("%d counters after an artifact died, want 2", got)
	}
	effectstest.ExileCreature(t, g, rock2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, dragon); got != 4 {
		t.Fatalf("%d counters after an artifact was exiled, want 4", got)
	}
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, dragon); got != 4 {
		t.Errorf("%d counters after a nonartifact died, want 4", got)
	}
}

// --- Tomb Blade ----------------------------------------------------------

func gapB189TombBladeHit(t *testing.T, victimCreatures int) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := effectstest.B12Push(g, me.ID, "Tomb Blade", gapB189ArtifactCreatureTL, gapB189TombBladeOracle, 5, 4)
	var ids []uuid.UUID
	for i := 0; i < victimCreatures; i++ {
		ids = append(ids, effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 0, 5))
	}
	effectstest.AttackWith(t, g, opp.ID, blade)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	return g, opp, ids
}

func TestGapB189TombBladeVictimMayLoseLifeEqualToTheirCreatures(t *testing.T) {
	g, opp, _ := gapB189TombBladeHit(t, 3)
	life := opp.Life
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 lost for 3 creatures", life, opp.Life)
	}
}

func TestGapB189TombBladeVictimMaySacrificeACreatureInstead(t *testing.T) {
	g, opp, ids := gapB189TombBladeHit(t, 3)
	life := opp.Life
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	sac := effectstest.PendingChooseCards(t, g, opp.ID)
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d, want no loss after sacrificing", life, opp.Life)
	}
	if g.Battlefield.Contains(ids[0]) {
		t.Error("the chosen creature was not sacrificed")
	}
}

func TestGapB189TombBladeAsksNothingOfAPlayerWithNoCreatures(t *testing.T) {
	g, opp, _ := gapB189TombBladeHit(t, 0)
	if effectstest.LatestOptionPickFor(g, opp.ID) != nil {
		t.Error("a player with no creatures was asked")
	}
}

// --- Technomancer ----------------------------------------------------------

func TestGapB189TechnomancerReturnsArtifactCreaturesUpToSixTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Top three, in the order they are milled: four, three and two mana value.
	effectstest.SeedLibraryTop(me, "Plain Bear", "Creature — Bear", "{1}")
	b := effectstest.SeedLibraryTop(me, "Bot Two", gapB189ArtifactCreatureTL2, "{2}")
	a := effectstest.SeedLibraryTop(me, "Bot Four", gapB189ArtifactCreatureTL2, "{4}")
	c := effectstest.SeedLibraryTop(me, "Bot Three", gapB189ArtifactCreatureTL2, "{3}")
	effectstest.AvB16CastCreature(t, g, "Technomancer", gapB189ArtifactCreatureTL, gapB189TechnomancerOracle, 5, 1)
	prompt := effectstest.PendingChooseCards(t, g, me.ID)
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a, c}); err == nil {
		t.Fatal("a pick with total mana value 7 was accepted")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("a pick with total mana value 6 was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneBattlefield {
			t.Errorf("a picked card is in %v, want the battlefield", effectstest.B12ZoneOf(g, id))
		}
	}
	if effectstest.B12ZoneOf(g, c) != game.ZoneGraveyard {
		t.Errorf("the unpicked card is in %v, want the graveyard", effectstest.B12ZoneOf(g, c))
	}
}

func TestGapB189TechnomancerOffersNothingWithoutArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.SeedLibraryTop(me, "Plain Bear", "Creature — Bear", "{1}")
	}
	effectstest.AvB16CastCreature(t, g, "Technomancer", gapB189ArtifactCreatureTL, gapB189TechnomancerOracle, 5, 1)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			t.Error("a prompt was opened with no artifact creature card in the graveyard")
		}
	}
}

// --- Szarekh, the Silent King ----------------------------------------------

func TestGapB189SzarekhTakesAnArtifactCreatureOrVehicleFromTheMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	szarekh := effectstest.B12Push(g, me.ID, "Szarekh, the Silent King", "Legendary "+gapB189ArtifactCreatureTL, gapB189SzarekhOracle, 3, 4)
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	veh := effectstest.SeedLibraryTop(me, "Buggy", "Artifact — Vehicle", "{2}")
	bot := effectstest.SeedLibraryTop(me, "Bot", gapB189ArtifactCreatureTL2, "{3}")
	effectstest.AttackWithTrigger(t, g, szarekh, opp.ID, uuid.Nil)
	prompt := effectstest.PendingChooseCards(t, g, me.ID)
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land card was accepted")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{veh}); err != nil {
		t.Fatal(err)
	}
	if effectstest.B12ZoneOf(g, veh) != game.ZoneHand {
		t.Errorf("the Vehicle is in %v, want the hand", effectstest.B12ZoneOf(g, veh))
	}
	if effectstest.B12ZoneOf(g, bot) != game.ZoneGraveyard {
		t.Errorf("the other card is in %v, want the graveyard", effectstest.B12ZoneOf(g, bot))
	}
}

func TestGapB189SzarekhOffersNothingWhenNothingQualifiesIsMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	szarekh := effectstest.B12Push(g, me.ID, "Szarekh, the Silent King", "Legendary "+gapB189ArtifactCreatureTL, gapB189SzarekhOracle, 3, 4)
	for i := 0; i < 3; i++ {
		effectstest.SeedLibraryTop(me, "Plain Bear", "Creature — Bear", "{1}")
	}
	effectstest.AttackWithTrigger(t, g, szarekh, opp.ID, uuid.Nil)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); p != nil {
		t.Error("a take prompt was opened with nothing to take")
	}
}

// --- Sloppity Bilepiper ----------------------------------------------------

func TestGapB189BilepiperGivesTheNextCreatureSpellCascadeOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pipe := effectstest.B12Push(g, me.ID, "Sloppity Bilepiper", "Creature — Demon", gapB189BilepiperOracle, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.SeedLibraryTop(me, "Cheap Spell", "Sorcery", "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pipe, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	// A noncreature spell does not use the ability up.
	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}"}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountMayCastPrompts(g); n != 0 {
		t.Fatalf("%d cascade offers after a noncreature spell", n)
	}
	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountMayCastPrompts(g); n != 1 {
		t.Fatalf("%d cascade offers after the creature spell, want 1", n)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "Second Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountMayCastPrompts(g); n != 0 {
		t.Errorf("%d cascade offers after a second creature spell, want 0", n)
	}
}

func TestGapB189BilepiperNeedsACreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pipe := effectstest.B12Push(g, me.ID, "Sloppity Bilepiper", "Creature — Demon", gapB189BilepiperOracle, 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pipe, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err == nil {
		t.Error("a noncreature permanent was accepted as the sacrifice")
	}
}
