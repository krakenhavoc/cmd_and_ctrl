package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b730_cards_test.go — Urza's Saga, batch 730 (#5624). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b730FlukeOracle       = "3c880ac2-0b22-4f7e-a55a-381de235e971"
	b730OwlOracle         = "7fa1eca4-0db1-446b-b3cb-15a619ffa97b"
	b730AetherOracle      = "a61ceda1-5993-479e-945f-15753eeb7049"
	b730WhaleOracle       = "c43ce30a-0346-479d-8dbc-56c3872ff82a"
	b730ResearchersOracle = "df2d2c52-475b-4057-9cf8-7886700f0fae"
	b730LifelineOracle    = "f475b636-eb7f-4dd7-bc87-a1af07064ffe"
	b730IndomitableOracle = "069c3c87-81a8-4d2a-ad2f-8d3243d0e3db"
	b730CreatureLine      = "Creature — Test"
)

func TestOldB730EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b730FlukeOracle:       "Spined Fluke",
		b730OwlOracle:         "Spire Owl",
		b730AetherOracle:      "Tainted Aether",
		b730WhaleOracle:       "Great Whale",
		b730ResearchersOracle: "Academy Researchers",
		b730LifelineOracle:    "Lifeline",
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

func b730OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func TestOldB730SpinedFlukeSacrificesACreatureOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", b730CreatureLine, 1, 1)
	fluke := effectstest.CastCatalogSpell(t, g, "Spined Fluke", "Creature — Worm Horror", b730FlukeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, fodder) || !b730OnBattlefield(g, fluke) {
		t.Error("the Fluke's trigger did not sacrifice the creature chosen")
	}
}

func TestOldB730SpinedFlukeAloneSacrificesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fluke := effectstest.CastCatalogSpell(t, g, "Spined Fluke", "Creature — Worm Horror", b730FlukeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fluke)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, fluke) {
		t.Error("the lone Fluke survived its own trigger")
	}
}

func TestOldB730SpinedFlukeRegeneratesForB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fluke := effectstest.PushCatalogPermanent(g, me.ID, "Spined Fluke", "Creature — Worm Horror", b730FlukeOracle, false)
	spec, _ := effects.Lookup(b730FlukeOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "{B}: Regenerate this creature." {
		t.Fatalf("activated abilities %v, want the {B} regeneration", spec.Activated)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, fluke, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B14Kill(g, fluke)
	if !b730OnBattlefield(g, fluke) {
		t.Error("the Fluke was destroyed despite its regeneration shield")
	}
}

func TestOldB730SpireOwlLooksAtTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Card", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastHoldCreature(t, g, "Spire Owl", "Creature — Bird", b730OwlOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-at prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Errorf("looking at %d cards, want 4", len(c.ScryCards))
	}
}

func TestOldB730GreatWhaleUntapsUpToSevenLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0 := g.Seats[g.Turn.ActiveSeat]
	lands := make([]uuid.UUID, 8)
	for i := range lands {
		lands[i] = effectstest.AangPushLand(g, p0.ID, "Island", true)
	}
	effectstest.CastCatalogSpell(t, g, "Great Whale", "Creature — Whale", b730WhaleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, p0.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 7 {
		t.Errorf("bounds %d..%d, want 0..7 (\"up to seven\")", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, p0.ID, lands[:7]...)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range lands {
		c, _ := effectstest.AangCardOnBF(g, id)
		if want := i >= 7; c.Tapped != want {
			t.Errorf("land %d tapped=%v, want %v", i, c.Tapped, want)
		}
	}
}

func TestOldB730TaintedAetherMakesTheNewcomersControllerSacrificeACreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.AangPushLand(g, opp.ID, "Swamp", false)
	other := effectstest.B12Creature(g, opp.ID, "Their Other", b730CreatureLine, 1, 1)
	mineLand := effectstest.AangPushLand(g, me.ID, "Forest", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Tainted Aether", "Enchantment", b730AetherOracle, false)
	newcomer := effectstest.DcEnter(t, g, opp.ID, "Newcomer", b730CreatureLine, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the creature's controller was not asked to sacrifice")
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the Aether's controller was asked, not the newcomer's")
	}
	if !effectstest.EdictOffers(effectstest.SacrificeOptions(c), newcomer) || !effectstest.EdictOffers(effectstest.SacrificeOptions(c), land) || !effectstest.EdictOffers(effectstest.SacrificeOptions(c), other) {
		t.Errorf("offers %v: want the newcomer, their other creature and their land", effectstest.SacrificeOptions(c))
	}
	if effectstest.EdictOffers(effectstest.SacrificeOptions(c), mineLand) {
		t.Error("another player's land was offered")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, land) {
		t.Error("the chosen land was not sacrificed")
	}
	if !b730OnBattlefield(g, newcomer) || !b730OnBattlefield(g, other) || !b730OnBattlefield(g, mineLand) {
		t.Error("something other than the chosen land was sacrificed")
	}
}

func TestOldB730TaintedAetherIgnoresANonCreatureEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tainted Aether", "Enchantment", b730AetherOracle, false)
	land := effectstest.AangPushLand(g, me.ID, "Forest", false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil || !b730OnBattlefield(g, land) {
		t.Error("a land entering triggered the Aether")
	}
}

func TestOldB730AcademyResearchersPutsAnAuraFromHandOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aura := effectstest.HandCardForTest(me, "Indomitable Might", "Enchantment — Aura", b730IndomitableOracle)
	notAura := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	res := effectstest.CastCatalogSpell(t, g, "Academy Researchers", "Creature — Human Wizard", b730ResearchersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the Aura")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may\", one card)", pick.ChooseMin, pick.ChooseMax)
	}
	if !effectstest.HasID(pick.ChooseCards, aura) || effectstest.HasID(pick.ChooseCards, notAura) {
		t.Errorf("offered %v: want only the Aura", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.FindBattlefieldCardForTest(g, aura)
	if got == nil || got.AttachedTo.ID != res {
		t.Errorf("the Aura is %v, want on the battlefield attached to the Researchers", got)
	}
}

func TestOldB730AcademyResearchersMayDeclineAndNothingIsAskedWithoutAnAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aura := effectstest.HandCardForTest(me, "Indomitable Might", "Enchantment — Aura", b730IndomitableOracle)
	effectstest.CastCatalogSpell(t, g, "Academy Researchers", "Creature — Human Wizard", b730ResearchersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, aura) {
		t.Error("an Aura entered though none was chosen")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g2, "Academy Researchers", "Creature — Human Wizard", b730ResearchersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, me2.ID) != nil {
		t.Error("asked for an Aura with none in hand")
	}
}

// b730Lifeline sets up a Lifeline, a creature of mine that will die and one
// that will not, and kills the first.
func b730LifelineScene(t *testing.T, withBystander bool) (g *game.Game, me *game.Player, dead, bystander uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Lifeline", "Artifact", b730LifelineOracle, false)
	dead = effectstest.B12Creature(g, me.ID, "Doomed", b730CreatureLine, 2, 2)
	if withBystander {
		bystander = effectstest.B12Creature(g, g.Seats[1].ID, "Bystander", b730CreatureLine, 1, 1)
	}
	effectstest.B14Kill(g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, dead, bystander
}

func TestOldB730LifelineReturnsTheCardAtTheNextEndStepUnderItsOwner(t *testing.T) {
	g, _, dead, _ := b730LifelineScene(t, true)
	if b730OnBattlefield(g, dead) {
		t.Fatal("the creature came back at once; it returns at the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !b730OnBattlefield(g, dead) {
		t.Error("the creature did not return at the beginning of the end step")
	}
}

func TestOldB730LifelineReturnsAnOpponentsCreatureUnderItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Lifeline", "Artifact", b730LifelineOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b730CreatureLine, 2, 2)
	effectstest.B12Creature(g, me.ID, "Mine", b730CreatureLine, 1, 1)
	effectstest.B14Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.FindBattlefieldCardForTest(g, theirs)
	if got == nil || got.Controller != opp.ID {
		t.Errorf("their creature is %v, want back under its owner's control", got)
	}
}

func TestOldB730LifelineDoesNothingWithNoOtherCreature(t *testing.T) {
	g, _, dead, _ := b730LifelineScene(t, false)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, dead) {
		t.Error("the creature returned though no other creature was on the battlefield")
	}
}

func TestOldB730LifelineStopsIfTheOtherCreatureIsGoneAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Lifeline", "Artifact", b730LifelineOracle, false)
	dead := effectstest.B12Creature(g, me.ID, "Doomed", b730CreatureLine, 2, 2)
	other := effectstest.B12Creature(g, g.Seats[1].ID, "Other", b730CreatureLine, 1, 1)
	effectstest.B14Kill(g, dead)
	// The trigger is on the stack; the other creature goes before it resolves.
	effectstest.B14Kill(g, other)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, dead) {
		t.Error("the creature returned though no other creature remained as the trigger resolved")
	}
}

func TestOldB730LifelineDoesNotReturnATokenOrACardThatLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Lifeline", "Artifact", b730LifelineOracle, false)
	dead := effectstest.B12Creature(g, me.ID, "Doomed", b730CreatureLine, 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Bystander", b730CreatureLine, 1, 1)
	effectstest.B14Kill(g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	// The card is exiled from the graveyard before the end step.
	g.WithWriteLock(func() {})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, dead); err != nil {
		t.Fatalf("exile from the graveyard: %v", err)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if b730OnBattlefield(g, dead) {
		t.Error("a card that left the graveyard came back")
	}
}
