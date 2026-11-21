package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b26_cards_test.go — Standard-set batch 26 (Duskmourn: House of
// Horror, #4466). Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	stdB26NashiOracle     = "ade8fc6f-efaa-4a24-96d8-99ef5026cd74"
	stdB26NikoOracle      = "f153e369-1d5b-4d8b-a2f6-d93b849cabb5"
	stdB26SprinterOracle  = "3416ac84-c5ef-44be-894f-ab3b89e592ce"
	stdB26BoxOracle       = "f0fd78bb-cb83-4dfa-a910-0334d225f87c"
	stdB26VacuumOracle    = "4ef583ab-e0a0-4e4b-bd65-af766038b1fd"
	stdB26RoleModelOracle = "8d24c12a-a300-4199-b2a9-3d5ead8df617"
	stdB26ComeBackOracle  = "2a9783c5-084b-4cd3-b7fa-ae9d7fed51d1"
	stdB26ChainsawOracle  = "3db40361-5f55-417e-a7cd-7e360cc91b4d"
	stdB26RecordingOracle = "9adc3750-0ddb-4f2c-a0e7-131e59d4cba5"
	stdB26LegendaryCreat  = "Legendary Creature — Human"
	stdB26EnchantmentType = "Enchantment"
	stdB26ZombieTypeLine  = "Creature — Zombie"
	stdB26BearTypeLine    = "Creature — Bear"
)

func TestStdB26EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB26NashiOracle:     "Nashi, Searcher in the Dark",
		stdB26NikoOracle:      "Niko, Light of Hope",
		stdB26SprinterOracle:  "Undead Sprinter",
		stdB26BoxOracle:       "Attack-in-the-Box",
		stdB26VacuumOracle:    "Ghost Vacuum",
		stdB26RoleModelOracle: "Reluctant Role Model",
		stdB26ComeBackOracle:  "Come Back Wrong",
		stdB26ChainsawOracle:  "Chainsaw",
		stdB26RecordingOracle: "Cursed Recording",
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

// stdB26Gy puts a card with a type line into p's graveyard.
func stdB26Gy(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// stdB26Counters reads a counter kind off a card wherever it is.
func stdB26Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return effectstest.B12Card(t, g, id).Counters[kind]
}

// --- Nashi, Searcher in the Dark ---------------------------------------------

// stdB26NashiTable has Nashi attacking an opponent with a legendary creature
// card, an enchantment card and a land on top of the library (the first two
// are milled by a 2-power hit, then the land). It returns the cards.
func stdB26NashiHit(t *testing.T, g *game.Game, me, opp *game.Player, nashi uuid.UUID) {
	t.Helper()
	effectstest.AttackWith(t, g, opp.ID, nashi)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestStdB26NashiTakesLegendaryAndEnchantmentCardsAndGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nashi := effectstest.B27Push(g, me.ID, "Nashi, Searcher in the Dark", "Legendary Creature — Rat Ninja Wizard", stdB26NashiOracle, "{U}{B}", 2, 2, "U", "B")
	land := effectstest.PlTop(me, "Island", "Basic Land — Island", "")
	ench := effectstest.PlTop(me, "Pacifism", stdB26EnchantmentType, "{1}{W}")
	legend := effectstest.PlTop(me, "Hero", stdB26LegendaryCreat, "{2}")
	_ = land

	stdB26NashiHit(t, g, me, opp, nashi)
	if opp.Life != 40-2 {
		t.Fatalf("defender life %d, want 38", opp.Life)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice among the milled cards")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, ench) || !effectstest.HasID(pick.ChooseCards, legend) {
		t.Fatalf("offered %v, want the legendary and the enchantment card only", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("pick range %d..%d, want any number (0..2)", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ench, legend)
	if !effectstest.MiInHand(me, ench) || !effectstest.MiInHand(me, legend) {
		t.Fatal("the chosen cards are not in hand")
	}
	if n := stdB26Counters(t, g, nashi, game.CounterPlusOne); n != 0 {
		t.Errorf("Nashi got %d counters although cards were put into hand", n)
	}
}

func TestStdB26NashiDecliningEveryCardPutsACounterOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nashi := effectstest.B27Push(g, me.ID, "Nashi, Searcher in the Dark", "Legendary Creature — Rat Ninja Wizard", stdB26NashiOracle, "{U}{B}", 2, 2, "U", "B")
	effectstest.PlTop(me, "Island", "Basic Land — Island", "")
	legend := effectstest.PlTop(me, "Hero", stdB26LegendaryCreat, "{2}")

	stdB26NashiHit(t, g, me, opp, nashi)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no choice among the milled cards")
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.MiInHand(me, legend) {
		t.Error("a declined card went to hand")
	}
	if n := stdB26Counters(t, g, nashi, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d after taking nothing, want 1", n)
	}
}

func TestStdB26NashiWithNothingToTakeAsksNothingAndGetsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nashi := effectstest.B27Push(g, me.ID, "Nashi, Searcher in the Dark", "Legendary Creature — Rat Ninja Wizard", stdB26NashiOracle, "{U}{B}", 2, 2, "U", "B")
	effectstest.PlTop(me, "Island", "Basic Land — Island", "")
	effectstest.PlTop(me, "Grizzly Bears", stdB26BearTypeLine, "{1}{G}")

	stdB26NashiHit(t, g, me, opp, nashi)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a choice was asked with no legendary or enchantment card milled")
	}
	if n := stdB26Counters(t, g, nashi, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
}

func TestStdB26NashiMillsOnlyAsManyAsTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nashi := effectstest.B27Push(g, me.ID, "Nashi, Searcher in the Dark", "Legendary Creature — Rat Ninja Wizard", stdB26NashiOracle, "{U}{B}", 2, 2, "U", "B")
	effectstest.PlTop(me, "Hero", stdB26LegendaryCreat, "{2}")
	effectstest.PlTop(me, "Island", "Basic Land — Island", "")
	effectstest.PlTop(me, "Island", "Basic Land — Island", "")

	before := me.Graveyard.Size()
	stdB26NashiHit(t, g, me, opp, nashi)
	if got := me.Graveyard.Size() - before; got != 2 {
		t.Errorf("milled %d cards, want 2 (the damage dealt)", got)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the legendary card below the milled two was offered")
	}
}

// --- Undead Sprinter ----------------------------------------------------------

func TestStdB26SprinterNeedsANonZombieCreatureToHaveDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sprinter := effectstest.TrekB13Card(g, me.Graveyard, me, "Undead Sprinter", "Creature — Zombie", "{B}{R}", stdB26SprinterOracle, 2, 2)

	effectstest.FloatMana(t, g, me, "{B}{R}")
	if err := g.CastSpell(me.ID, sprinter, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard with no creature having died")
	}

	zombie := effectstest.B12Creature(g, me.ID, "Walking Corpse", stdB26ZombieTypeLine, 2, 2)
	effectstest.B27Kill(g, zombie)
	g.RunStateChecksForTest()
	if err := g.CastSpell(me.ID, sprinter, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard after only a Zombie died")
	}

	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	if err := g.CastSpell(me.ID, sprinter, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard after a Bear died: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, sprinter, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d on a Sprinter cast from the graveyard, want 1", n)
	}
	if p := effectstest.LayeredCard(t, g, sprinter).CurrentPower(); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	abilities := effectstest.EffectiveAbilities(t, g, sprinter)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasString(abilities, kw) {
			t.Errorf("Undead Sprinter is missing %q: %v", kw, abilities)
		}
	}
}

func TestStdB26SprinterAChangelingDeathDoesNotCountAndAHandCastGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sprinter := effectstest.TrekB13Card(g, me.Graveyard, me, "Undead Sprinter", "Creature — Zombie", "{B}{R}", stdB26SprinterOracle, 2, 2)
	changeling := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, Keywords: []string{game.KeywordChangeling},
	})
	effectstest.B27Kill(g, changeling)
	g.RunStateChecksForTest()
	effectstest.FloatMana(t, g, me, "{B}{R}")
	if err := g.CastSpell(me.ID, sprinter, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("a changeling is a Zombie, so its death does not open the graveyard")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	hand := effectstest.TrekB13Card(g2, me2.Hand, me2, "Undead Sprinter", "Creature — Zombie", "{B}{R}", stdB26SprinterOracle, 2, 2)
	effectstest.FloatMana(t, g2, me2, "{B}{R}")
	if err := g2.CastSpell(me2.ID, hand, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if n := stdB26Counters(t, g2, hand, game.CounterPlusOne); n != 0 {
		t.Errorf("+1/+1 counters %d on a Sprinter cast from hand, want 0", n)
	}
}

// --- Attack-in-the-Box --------------------------------------------------------

func stdB26BoxAttack(t *testing.T, g *game.Game, me, opp *game.Player, yes bool) uuid.UUID {
	t.Helper()
	box := effectstest.B27Push(g, me.ID, "Attack-in-the-Box", "Artifact Creature — Toy", stdB26BoxOracle, "{3}", 2, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(box, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
	effectstest.PassPriorityAroundTable(t, g)
	return box
}

func TestStdB26BoxPumpedAttackHitsForSixAndIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := stdB26BoxAttack(t, g, me, opp, true)
	if p := effectstest.EffectivePower(t, g, box); p != 6 {
		t.Fatalf("power %d after saying yes, want 6", p)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != 40-6 {
		t.Errorf("defender life %d, want 34", opp.Life)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.B12ZoneOf(g, box) != game.ZoneBattlefield {
		t.Fatal("the box was gone before the end step trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, box) != game.ZoneGraveyard {
		t.Errorf("the box is in %q at the end step, want the graveyard", effectstest.B12ZoneOf(g, box))
	}
}

func TestStdB26BoxDecliningKeepsItSmallAndAlive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := stdB26BoxAttack(t, g, me, opp, false)
	if p := effectstest.EffectivePower(t, g, box); p != 2 {
		t.Fatalf("power %d after saying no, want 2", p)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, box) != game.ZoneBattlefield {
		t.Errorf("the box is in %q after declining, want the battlefield", effectstest.B12ZoneOf(g, box))
	}
}
