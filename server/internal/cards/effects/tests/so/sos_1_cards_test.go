package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sos_1_cards_test.go — Secrets of Strixhaven, first batch (#4456). Each
// test plays the real catalog entry and asserts what its text says.

const (
	sos1AbundanceOracle   = cardse.EmeritusOfAbundanceOracleID
	sos1ConflictOracle    = cardse.EmeritusOfConflictOracleID
	sos1TruceOracle       = cardse.EmeritusOfTruceOracleID
	sos1WoeOracle         = cardse.EmeritusOfWoeOracleID
	sos1GerminationOracle = "c8d22139-7b25-4bb1-b0d9-d4c622401edf"
	sos1SeminarOracle     = "fd0a27a9-7d7e-4a16-9dd0-0715e6c97b36"
	sos1WitheringOracle   = "fddd1c39-bdf8-4aee-897c-666498cc4dd3"
	sos1MathemagicsOracle = "5f56a303-b3da-41df-a959-7823a87ee60c"
	sos1TrioOracle        = cardsh.HarmonizedTrioOracleID
	sos1ArtisanOracle     = cardslu.MaelstromArtisanOracleID
	sos1SplatterOracle    = "2620dcfe-8738-4f81-b23d-0c1c60cf9047"
	sos1TogetherOracle    = "234f1c4b-dd0c-4c4e-b3c1-3b9ccfc42226"
	sos1CounselOracle     = "5fd193c0-ed9f-4afb-b6cd-c1a20008fc0b"
	sos1CritiqueOracle    = "06c1f8a4-4948-435b-ae98-b1276b57e958"
	sos1SanarOracle       = effects.SanarUnfinishedGeniusOracleID
	sos1BlechOracle       = "c1bdd361-c7f2-4aa2-8d07-b582775551ff"
	sos1ResearchersOracle = cardsj.JoinedResearchersOracleID
	sos1RevitalizeOracle  = "b1385b03-cb4b-4812-857f-7421f1df39af"
)

func TestSos1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		sos1AbundanceOracle:          "Emeritus of Abundance",
		sos1AbundanceOracle + "#1":   "Regrowth",
		sos1ConflictOracle:           "Emeritus of Conflict",
		sos1ConflictOracle + "#1":    "Lightning Bolt",
		sos1TruceOracle:              "Emeritus of Truce",
		sos1TruceOracle + "#1":       "Swords to Plowshares",
		sos1WoeOracle:                "Emeritus of Woe",
		sos1WoeOracle + "#1":         "Demonic Tutor",
		sos1GerminationOracle:        "Germination Practicum",
		sos1SeminarOracle:            "Restoration Seminar",
		sos1WitheringOracle:          "Withering Curse",
		sos1MathemagicsOracle:        "Mathemagics",
		sos1TrioOracle:               "Harmonized Trio",
		sos1TrioOracle + "#1":        "Brainstorm",
		sos1ArtisanOracle:            "Maelstrom Artisan",
		sos1ArtisanOracle + "#1":     "Rocket Volley",
		sos1SplatterOracle:           "Splatter Technique",
		sos1TogetherOracle:           "Together as One",
		sos1CounselOracle:            "Comforting Counsel",
		sos1CritiqueOracle:           "Traumatic Critique",
		sos1SanarOracle:              "Sanar, Unfinished Genius",
		sos1SanarOracle + "#1":       "Wild Idea",
		sos1BlechOracle:              "Blech, Loafing Pest",
		sos1ResearchersOracle:        "Joined Researchers",
		sos1ResearchersOracle + "#1": "Secret Rendezvous",
	}
	for oracle, name := range want {
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

// sos1Prep builds a preparation card in the shape deck import gives.
func sos1Prep(owner uuid.UUID, oracle string, keywords []string, name, typeLine, manaCost string, power, toughness int, spell, spellType, spellCost string) game.Card {
	return effectstest.PreparationCard(owner, oracle, keywords,
		game.Face{Name: name, TypeLine: typeLine, ManaCost: manaCost, Power: power, Toughness: toughness},
		game.Face{Name: spell, TypeLine: spellType, ManaCost: spellCost})
}

// sos1Lands puts n basic Forests onto the battlefield under owner.
func sos1Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Forest", "Basic Land — Forest")
	}
}

// sos1GainLife casts Revitalize ("You gain 3 life. Draw a card.") for the
// active seat and lets the table settle, so a "whenever you gain life"
// trigger is raised and resolved the way play raises it.
func sos1GainLife(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Revitalize", "Instant", sos1RevitalizeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Emeritus of Abundance // Regrowth -----------------------------------

func TestSos1EmeritusOfAbundanceEntersPreparedAndRegrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, sos1AbundanceOracle, []string{"vigilance"}, "Emeritus of Abundance", "Creature — Elf Druid", "{2}{G}", 3, 4,
		"Regrowth", "Sorcery", "{1}{G}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Emeritus of Abundance did not enter prepared")
	}
	gone := effectstest.PushGraveyardCardForTest(me, "Lost Card")
	effectstest.FraBCastCopy(t, g, me, "Regrowth", game.TargetRef{Kind: game.TargetCard, ID: gone})
	if !me.Hand.Contains(gone) {
		t.Error("Regrowth did not return the card to hand")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSos1EmeritusOfAbundanceAttackPreparesOnlyWithEightLands(t *testing.T) {
	for _, tc := range []struct {
		lands int
		want  bool
	}{{7, false}, {8, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := sos1Prep(me.ID, sos1AbundanceOracle, []string{"vigilance"}, "Emeritus of Abundance", "Creature — Elf Druid", "{2}{G}", 3, 4,
			"Regrowth", "Sorcery", "{1}{G}")
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
		sos1Lands(g, me.ID, tc.lands)
		if g.IsPreparedForEffect(c.InstanceID) {
			t.Fatal("put onto the battlefield, it should not be prepared yet")
		}
		effectstest.DeclareAttack(t, g, opp.ID, c.InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.IsPreparedForEffect(c.InstanceID); got != tc.want {
			t.Errorf("%d lands: prepared = %v, want %v", tc.lands, got, tc.want)
		}
	}
}

// --- Emeritus of Conflict // Lightning Bolt ------------------------------

func TestSos1EmeritusOfConflictPreparesOnTheThirdSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := sos1Prep(me.ID, sos1ConflictOracle, []string{"first strike"}, "Emeritus of Conflict", "Creature — Human Wizard", "{1}{R}", 2, 2,
		"Lightning Bolt", "Instant", "{R}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	for i := 1; i <= 3; i++ {
		effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got, want := g.IsPreparedForEffect(c.InstanceID), i == 3; got != want {
			t.Fatalf("after spell %d: prepared = %v, want %v", i, got, want)
		}
	}
	life := opp.Life
	effectstest.FraBCastCopy(t, g, me, "Lightning Bolt", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-3 {
		t.Errorf("Lightning Bolt: life %d -> %d, want -3", life, opp.Life)
	}
	// The copy was the fourth spell: it does not prepare it again.
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("the fourth spell prepared it again")
	}
}

// --- Emeritus of Truce // Swords to Plowshares ---------------------------

func TestSos1EmeritusOfTruceGivesTheInklingAndComparesAfter(t *testing.T) {
	for _, tc := range []struct {
		name    string
		toOpp   bool
		prepare bool
	}{{"Inkling to an opponent with a creature", true, true}, {"Inkling kept", false, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			c := sos1Prep(me.ID, sos1TruceOracle, nil, "Emeritus of Truce", "Creature — Cat Cleric", "{1}{W}{W}", 3, 3,
				"Swords to Plowshares", "Instant", "{W}")
			effectstest.CastFromHandAndResolve(t, g, me, c)
			who := me.ID
			if tc.toOpp {
				who = opp.ID
			}
			effectstest.PickPlayer(t, g, me.ID, who)
			effectstest.PassPriorityAroundTable(t, g)
			inklings := 0
			for _, card := range g.Battlefield.Cards {
				if card.Name == "Inkling" && card.Controller == who && card.Power == 1 && card.Toughness == 1 {
					inklings++
				}
			}
			if inklings != 1 {
				t.Fatalf("%d 1/1 Inklings under the target player, want 1", inklings)
			}
			if got := g.IsPreparedForEffect(c.InstanceID); got != tc.prepare {
				t.Errorf("prepared = %v, want %v", got, tc.prepare)
			}
		})
	}
}

func TestSos1EmeritusOfTruceSwordsExilesAndPaysTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := sos1Prep(me.ID, sos1TruceOracle, nil, "Emeritus of Truce", "Creature — Cat Cleric", "{1}{W}{W}", 3, 3,
		"Swords to Plowshares", "Instant", "{W}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	if _, err := g.BecomePreparedForEffect(c.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4)
	life := opp.Life
	effectstest.FraBCastCopy(t, g, me, "Swords to Plowshares", game.TargetRef{Kind: game.TargetCard, ID: bear})
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	if opp.Life != life+4 {
		t.Errorf("its controller gains its power: %d -> %d, want +4", life, opp.Life)
	}
}

// --- Emeritus of Woe // Demonic Tutor ------------------------------------

func TestSos1EmeritusOfWoeEntersPreparedAndTutors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, sos1WoeOracle, nil, "Emeritus of Woe", "Creature — Vampire Warlock", "{3}{B}", 5, 4,
		"Demonic Tutor", "Sorcery", "{1}{B}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Emeritus of Woe did not enter prepared")
	}
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle"})
	effectstest.FraBCastCopy(t, g, me, "Demonic Tutor")
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("Demonic Tutor did not put the card into hand")
	}
}

func TestSos1EmeritusOfWoePreparesAtYourEndStepAfterTwoDeaths(t *testing.T) {
	for _, tc := range []struct {
		deaths int
		want   bool
	}{{1, false}, {2, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		c := sos1Prep(me.ID, sos1WoeOracle, nil, "Emeritus of Woe", "Creature — Vampire Warlock", "{3}{B}", 5, 4,
			"Demonic Tutor", "Sorcery", "{1}{B}")
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
		g.WithWriteLock(func() { g.TurnTally.CreaturesDied = tc.deaths })
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.IsPreparedForEffect(c.InstanceID); got != tc.want {
			t.Errorf("%d creatures died: prepared = %v, want %v", tc.deaths, got, tc.want)
		}
	}
}

// --- Germination Practicum / Restoration Seminar -------------------------

func TestSos1GerminationPracticumCountersYourCreaturesAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Germination Practicum", "Sorcery — Lesson", sos1GerminationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("your creature has %d counters, want 2", got)
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", got)
	}
	if me.Graveyard.Contains(id) {
		t.Error("paradigm: the spell went to the graveyard instead of exile")
	}
}

func TestSos1RestorationSeminarReturnsANonlandPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.Batch01GraveyardCard(me, "Relic", "Artifact")
	land := effectstest.Batch01GraveyardCard(me, "Dead Forest", "Land")
	if err := effectstest.CastCatalogSpellErr(t, g, "Restoration Seminar", "Sorcery — Lesson", sos1SeminarOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a land card as the target: err = %v, want ErrIllegalTarget", err)
	}
	effectstest.CastCatalogSpell(t, g, "Restoration Seminar", "Sorcery — Lesson", sos1SeminarOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("the artifact card did not return to the battlefield")
	}
}

// --- Withering Curse -----------------------------------------------------

func TestSos1WitheringCurseShrinksWithoutLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Withering Curse", "Sorcery", sos1WitheringOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 should die to -2/-2")
	}
	if !g.Battlefield.Contains(big) {
		t.Fatal("a 3/3 should survive -2/-2")
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the 3/3 is %d/%d, want 1/1", p, tough)
	}
}

func TestSos1WitheringCurseDestroysAllWithLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 5, 5)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.CastCatalogSpell(t, g, "Withering Curse", "Sorcery", sos1WitheringOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(theirs) {
		t.Error("with life gained this turn every creature is destroyed")
	}
}

// --- Mathemagics ---------------------------------------------------------

func TestSos1MathemagicsDrawsTwoToTheX(t *testing.T) {
	for _, tc := range []struct{ x, want int }{{0, 1}, {2, 4}, {3, 8}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		hand := opp.Hand.Size()
		effectstest.Pr8CastX(t, g, "Mathemagics", "Sorcery", sos1MathemagicsOracle, tc.x,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		if got := opp.Hand.Size() - hand; got != tc.want {
			t.Errorf("X = %d: drew %d, want %d", tc.x, got, tc.want)
		}
	}
	if got := cardslu.MathemagicsCount(64); got != game.MaxCount {
		t.Errorf("a huge X is pinned to the count ceiling: %d, want %d", got, game.MaxCount)
	}
}

// --- Harmonized Trio // Brainstorm ---------------------------------------

func TestSos1HarmonizedTrioTapsTwoOthersToPrepare(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, sos1TrioOracle, nil, "Harmonized Trio", "Creature — Merfolk Bard Wizard", "{U}", 1, 1,
		"Brainstorm", "Instant", "{U}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("one other creature paid a cost that needs two")
	}
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, c.InstanceID}}); err == nil {
		t.Fatal("the Trio paid for itself twice: {T} and one of the two")
	}
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("the Trio did not become prepared")
	}
	for _, id := range []uuid.UUID{c.InstanceID, a, b} {
		if !effectstest.B12Card(t, g, id).Tapped {
			t.Errorf("%s is untapped", effectstest.B12Card(t, g, id).Name)
		}
	}
	hand := me.Hand.Size()
	cp, ok := effectstest.PrepareCopyOf(g, "Brainstorm")
	if !ok {
		t.Fatal("no copy of Brainstorm in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast Brainstorm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("Brainstorm drew to %d, want %d before the put-back", got, hand+3)
	}
}

// --- Maelstrom Artisan // Rocket Volley ----------------------------------

func TestSos1MaelstromArtisanRocketVolleyHitsOnlyNonbasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, sos1ArtisanOracle, []string{"haste"}, "Maelstrom Artisan", "Creature — Minotaur Sorcerer", "{1}{R}{R}", 3, 2,
		"Rocket Volley", "Sorcery", "{1}{R}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Maelstrom Artisan did not enter prepared")
	}
	basic := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	shock := effectstest.B12Permanent(g, opp.ID, "Sacred Foundry", "Land — Mountain Plains")
	cp, _ := effectstest.PrepareCopyOf(g, "Rocket Volley")
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile),
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: basic}}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Rocket Volley on a basic land: err = %v, want ErrIllegalTarget", err)
	}
	effectstest.FraBCastCopy(t, g, me, "Rocket Volley", game.TargetRef{Kind: game.TargetCard, ID: shock})
	if g.Battlefield.Contains(shock) {
		t.Error("the nonbasic land was not destroyed")
	}
}

// --- Splatter Technique --------------------------------------------------

func TestSos1SplatterTechniqueDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Splatter Technique", "Sorcery", sos1SplatterOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The cast is a wash: the spell was pushed to hand and then cast.
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
}

func TestSos1SplatterTechniqueDamagesEachCreatureAndPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 4, 4)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	walker := effectstest.B12Permanent(g, opp.ID, "Walker", "Legendary Planeswalker — Test")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 6}
			}
		}
	})
	effectstest.CastModal(t, g, "Splatter Technique", "Sorcery", sos1SplatterOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("your own 4/4 should die too")
	}
	if !g.Battlefield.Contains(big) || effectstest.B12Card(t, g, big).DamageMarked != 4 {
		t.Errorf("the 5/5 should survive with 4 damage")
	}
	if got := effectstest.CounterCount(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("the planeswalker has %d loyalty, want 2", got)
	}
}

// --- Together as One -----------------------------------------------------

func TestSos1TogetherAsOneConvergesForTheColoursSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.FloatForTest(g, me, "WUBCCC")
	hand, life := opp.Hand.Size(), me.Life
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Together as One", TypeLine: "Sorcery", ManaCost: "{6}",
			OracleID: sos1TogetherOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 3 {
		t.Errorf("the target player drew %d, want 3 (three colours)", got)
	}
	if got := effectstest.B12Card(t, g, bear).DamageMarked; got != 3 {
		t.Errorf("the creature was dealt %d, want 3", got)
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("you gained %d, want 3", got)
	}
}

// --- Comforting Counsel --------------------------------------------------

func TestSos1ComfortingCounselTurnsOnAtFiveGrowthCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	counsel := effectstest.B12Push(g, me.ID, "Comforting Counsel", "Enchantment", sos1CounselOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for i := 1; i <= 5; i++ {
		sos1GainLife(t, g)
		if got := effectstest.CounterCount(g, counsel, "growth"); got != i {
			t.Fatalf("after %d life gains: %d growth counters", i, got)
		}
		wantPower := 2
		if i >= 5 {
			wantPower = 5
		}
		if got := effectstest.EffectivePower(t, g, mine); got != wantPower {
			t.Errorf("after %d gains your creature has power %d, want %d", i, got, wantPower)
		}
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", got)
	}
}

// --- Traumatic Critique --------------------------------------------------

func TestSos1TraumaticCritiqueDealsXThenLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	hand := me.Hand.Size()
	effectstest.Pr8CastX(t, g, "Traumatic Critique", "Instant", sos1CritiqueOracle, 3, effectstest.CardRefs(bear))
	if got := effectstest.B12Card(t, g, bear).DamageMarked; got != 3 {
		t.Errorf("the creature was dealt %d, want 3", got)
	}
	if c := effectstest.DiscardChoiceFor(g, me.ID); c == nil || c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Fatalf("draw two, then discard a card: discard prompt %+v", c)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	// The cast is a wash: +2 drawn, -1 discarded.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1", got)
	}
}

// --- Sanar, Unfinished Genius // Wild Idea -------------------------------

func TestSos1SanarMakesATreasureOnlyAfterAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := sos1Prep(me.ID, sos1SanarOracle, nil, "Sanar, Unfinished Genius", "Legendary Creature — Goblin Sorcerer", "{U}{R}", 0, 4,
		"Wild Idea", "Sorcery", "{3}{U}{R}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the Treasure ability activated before any instant or sorcery was cast")
	}
	effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate after an instant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure was created")
	}
}

func TestSos1SanarEntersPreparedAndWildIdeaCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, sos1SanarOracle, nil, "Sanar, Unfinished Genius", "Legendary Creature — Goblin Sorcerer", "{U}{R}", 0, 4,
		"Wild Idea", "Sorcery", "{3}{U}{R}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Sanar did not enter prepared")
	}
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Idea", TypeLine: "Instant"})
	effectstest.FraBCastCopy(t, g, me, "Wild Idea")
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Fatal("Wild Idea did not fetch the instant")
	}
	// The copy of Wild Idea was cast, and it has ceased to exist: it
	// still counts as an instant or sorcery cast this turn.
	if !effects.SanarCastAnInstantOrSorceryThisTurn(g, me.ID, c.InstanceID) {
		t.Error("casting the copy of Wild Idea did not count as casting a sorcery")
	}
}

// --- Blech, Loafing Pest -------------------------------------------------

func TestSos1BlechGrowsItsFiveTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blech := effectstest.B12Push(g, me.ID, "Blech, Loafing Pest", "Legendary Creature — Pest", sos1BlechOracle, 3, 4)
	insect := effectstest.B12Creature(g, me.ID, "Bug", "Creature — Insect", 1, 1)
	spider := effectstest.B12Creature(g, me.ID, "Spider", "Creature — Spider", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBat := effectstest.B12Creature(g, opp.ID, "Their Bat", "Creature — Bat", 1, 1)
	sos1GainLife(t, g)
	for _, id := range []uuid.UUID{blech, insect, spider} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d counters, want 1", effectstest.B12Card(t, g, id).Name, got)
		}
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a Bear has %d counters, want 0", got)
	}
	if got := effectstest.CounterCount(g, theirBat, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Bat has %d counters, want 0", got)
	}
}

// --- Joined Researchers // Secret Rendezvous -----------------------------

func TestSos1JoinedResearchersPreparesWhenAnOpponentHoldsMore(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra int
		want  bool
	}{{"even hands", 0, false}, {"an opponent holds more", 1, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			c := sos1Prep(me.ID, sos1ResearchersOracle, []string{"first strike"}, "Joined Researchers", "Creature — Human Cleric Wizard", "{1}{W}", 2, 2,
				"Secret Rendezvous", "Sorcery", "{1}{W}{W}")
			effectstest.PushBattlefieldCardWithTimestamp(g, c)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
			g.WithWriteLock(func() {
				for _, p := range g.Seats {
					p.Hand.Cards = nil
				}
				for i := 0; i < 2; i++ {
					me.Hand.PushTop(game.NewCard("Mine", me.ID))
					opp.Hand.PushTop(game.NewCard("Theirs", opp.ID))
				}
				for i := 0; i < tc.extra; i++ {
					opp.Hand.PushTop(game.NewCard("Extra", opp.ID))
				}
			})
			// "Each end step": your own counts, and it is asked as the
			// step begins.
			effectstest.AdvanceToEndStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := g.IsPreparedForEffect(c.InstanceID); got != tc.want {
				t.Errorf("prepared = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSos1SecretRendezvousCopyDrawsForBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, sos1ResearchersOracle, []string{"first strike"}, "Joined Researchers", "Creature — Human Cleric Wizard", "{1}{W}", 2, 2,
		"Secret Rendezvous", "Sorcery", "{1}{W}{W}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	if _, err := g.BecomePreparedForEffect(c.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.FraBCastCopy(t, g, me, "Secret Rendezvous", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if me.Hand.Size()-mine != 3 || opp.Hand.Size()-theirs != 3 {
		t.Errorf("draws: you %d, them %d, want 3 each", me.Hand.Size()-mine, opp.Hand.Size()-theirs)
	}
}
