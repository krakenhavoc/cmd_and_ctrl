package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 GAP batch 172 (Jumpstart 2022 instants, #5948).

const (
	gapB172SailsOracle    = "ea908d2d-d443-4385-a71a-ebb00938cd57"
	gapB172KickOracle     = "8c54fd75-e89f-490e-90dd-788f22b6e316"
	gapB172SparksOracle   = "82db3a7e-ec3f-4f25-996e-9e1c23c86003"
	gapB172EpiphanyOracle = "ed10bb4c-f7ef-4046-8dde-465041b55078"
	gapB172CondescendOrcl = "96cf3c10-733d-4110-9dab-43d0cd4e6629"
	gapB172ChillOracle    = "4459117a-d11c-44f4-9aac-caefc5b2d6f2"
	gapB172DispersalOrcl  = "e2e44d46-9df4-46f8-acf7-5581b2257620"
)

func TestGapB172CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB172SailsOracle: "Shredded Sails", gapB172KickOracle: "Swift Kick",
		gapB172SparksOracle: "Welding Sparks", gapB172EpiphanyOracle: "Artificer's Epiphany",
		gapB172CondescendOrcl: "Condescend", gapB172ChillOracle: "Crippling Chill",
		gapB172DispersalOrcl: "Lookout's Dispersal",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// gapB172Cast puts the card in the active seat's hand, funds `mana` and
// casts it, returning the cast error.
func gapB172Cast(t *testing.T, g *game.Game, card game.Card, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card.InstanceID, card.Owner, card.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(card)
	effectstest.AdvanceToMainForTest(t, g)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}

func TestGapB172ShreddedSailsDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if _, err := gapB172Cast(t, g, game.Card{Name: "Shredded Sails", TypeLine: "Instant", OracleID: gapB172SailsOracle, ManaCost: "{1}{R}"},
		"{C}{R}", game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, rock) {
		t.Error("the artifact survived")
	}
}

func TestGapB172ShreddedSailsHitsOnlyAFlyingCreatureForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 2, Toughness: 4,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	card := game.Card{Name: "Shredded Sails", TypeLine: "Instant", OracleID: gapB172SailsOracle, ManaCost: "{1}{R}"}
	if _, err := gapB172Cast(t, g, card, "{C}{R}", game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTarget(bear), Strict: true}); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	if _, err := gapB172Cast(t, g, card, "{C}{R}", game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTarget(bird)}); err != nil {
		t.Fatalf("cast at the flyer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, bird) {
		t.Error("a 2/4 flyer survived 4 damage")
	}
}

func TestGapB172ShreddedSailsCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Shredded Sails", "Instant", gapB172SailsOracle, "{C}{C}")
	if !effectstest.InGraveyard(me, id) {
		t.Error("the cycled card is not in the graveyard")
	}
}

func TestGapB172SwiftKickPumpsThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 3, 3)
	card := game.Card{Name: "Swift Kick", TypeLine: "Instant", OracleID: gapB172KickOracle, ManaCost: "{3}{R}"}
	if _, err := gapB172Cast(t, g, card, "{C}{C}{C}{R}", game.CastSpellParams{Targets: effectstest.CardTargets(mine, theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, theirs) {
		t.Error("the +1/+0 creature (3 power) did not kill the 3/3 in the fight")
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("the creature's power is %d, want 3 (2 +1/+0)", p)
	}
	if !effectstest.BattlefieldHas(g, mine) {
		t.Error("the 3/4 died to 3 damage")
	}
}

func TestGapB172SwiftKickRefusesAnOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	card := game.Card{Name: "Swift Kick", TypeLine: "Instant", OracleID: gapB172KickOracle, ManaCost: "{3}{R}"}
	if _, err := gapB172Cast(t, g, card, "{C}{C}{C}{R}", game.CastSpellParams{Targets: effectstest.CardTargets(a, b), Strict: true}); err == nil {
		t.Error("two creatures you control were accepted as the fight's targets")
	}
}

func TestGapB172WeldingSparksCountsYourArtifacts(t *testing.T) {
	for _, tc := range []struct {
		artifacts int
		dies      bool
	}{{0, false}, {2, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < tc.artifacts; i++ {
			effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
		}
		victim := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
		card := game.Card{Name: "Welding Sparks", TypeLine: "Instant", OracleID: gapB172SparksOracle, ManaCost: "{2}{R}"}
		if _, err := gapB172Cast(t, g, card, "{C}{C}{R}", game.CastSpellParams{Targets: effectstest.CardTarget(victim)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if died := !effectstest.BattlefieldHas(g, victim); died != tc.dies {
			t.Errorf("%d artifacts: the 0/5 died = %v, want %v", tc.artifacts, died, tc.dies)
		}
	}
}

func TestGapB172ArtificersEpiphanyDiscardsOnlyWithoutArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()
	card := game.Card{Name: "Artificer's Epiphany", TypeLine: "Instant", OracleID: gapB172EpiphanyOracle, ManaCost: "{2}{U}"}
	if _, err := gapB172Cast(t, g, card, "{C}{C}{U}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("with no artifact: hand %d, want %d (draw two, discard one)", got, before+1)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.SeedLibrary(me2, "A", "B", "C")
	effectstest.B12Permanent(g2, me2.ID, "Rock", "Artifact")
	before2 := me2.Hand.Size()
	if _, err := gapB172Cast(t, g2, card, "{C}{C}{U}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != before2+2 {
		t.Errorf("with an artifact: hand %d, want %d (draw two, no discard)", got, before2+2)
	}
	if effectstest.DiscardChoiceFor(g2, me2.ID) != nil {
		t.Error("a discard was asked of a player who controls an artifact")
	}
}

func TestGapB172CondescendCountersUnlessXIsPaidAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	card := game.Card{Name: "Condescend", TypeLine: "Instant", OracleID: gapB172CondescendOrcl, ManaCost: "{X}{U}"}
	if _, err := gapB172Cast(t, g, card, "{C}{C}{U}", game.CastSpellParams{Targets: effectstest.CardTarget(spell), XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt after the counter")
	}
}

func TestGapB172CripplingChillTapsHoldsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	card := game.Card{Name: "Crippling Chill", TypeLine: "Instant", OracleID: gapB172ChillOracle, ManaCost: "{2}{U}"}
	if _, err := gapB172Cast(t, g, card, "{C}{C}{U}", game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, before+1)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if g.Turn.ActiveSeat != 1 {
		t.Skipf("advanced to seat %d, not the creature's controller", g.Turn.ActiveSeat)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature untapped during its controller's next untap step")
	}
}

func TestGapB172LookoutsDispersalCostsOneLessWithAPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Corsair", "Creature — Human Pirate", 2, 2)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	card := game.Card{Name: "Lookout's Dispersal", TypeLine: "Instant", OracleID: gapB172DispersalOrcl, ManaCost: "{2}{U}"}
	// With a Pirate it costs {1}{U}: two mana is enough.
	if _, err := gapB172Cast(t, g, card, "{C}{U}", game.CastSpellParams{Targets: effectstest.CardTarget(spell), Strict: true}); err != nil {
		t.Fatalf("cast with a Pirate and two mana: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
}

func TestGapB172LookoutsDispersalCostsFullPriceWithoutAPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	card := game.Card{Name: "Lookout's Dispersal", TypeLine: "Instant", OracleID: gapB172DispersalOrcl, ManaCost: "{2}{U}"}
	if _, err := gapB172Cast(t, g, card, "{C}{U}", game.CastSpellParams{Targets: effectstest.CardTarget(spell), Strict: true}); err == nil {
		t.Error("two mana paid for a {2}{U} spell with no Pirate")
	}
}
