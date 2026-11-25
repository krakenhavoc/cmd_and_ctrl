package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB176Bankrupt  = "10a66403-606d-4814-9583-50cdcd6976e0"
	oldB176ClearMind = "141c33be-7a25-4981-b62c-efee8e8ec91d"
	oldB176Nodorog   = "3992209a-0a28-449d-8ba5-f75e861a79b4"
	oldB176Tome      = "3c7dd86d-25eb-4b24-9358-98b951248e04"
	oldB176Dismissal = "6db5f1a7-4cd8-42e2-a5d8-1603206ad0cc"
	oldB176Acuity    = "898e480d-c44e-4080-910a-f53e4ee471e1"
	oldB176Vannifar  = "b543bfa2-0174-4dd8-9ee6-09b92fb44934"
	oldB176Amplifire = "c6861c50-d336-4352-a76d-ec22a86ec09e"
)

func TestOldB176AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB176Bankrupt: "Bankrupt in Blood", oldB176ClearMind: "Clear the Mind", oldB176Nodorog: "Domri's Nodorog",
		oldB176Tome: "Tome of the Guildpact", oldB176Dismissal: "Dovin's Dismissal", oldB176Acuity: "Dovin's Acuity",
		oldB176Vannifar: "Prime Speaker Vannifar", oldB176Amplifire: "Amplifire",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB176Cast(t *testing.T, g *game.Game, me *game.Player, card game.Card, params game.CastSpellParams) error {
	t.Helper()
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, card.InstanceID, params)
}

func TestOldB176BankruptInBloodSacrificesTwoCreaturesToDrawThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.SeedLibrary(me, "x", "y", "z", "w")
	card := game.Card{Name: "Bankrupt in Blood", TypeLine: "Sorcery", OracleID: oldB176Bankrupt}
	if err := oldB176Cast(t, g, me, card, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("cast with one creature sacrificed")
	}
	hand := me.Hand.Size()
	if err := oldB176Cast(t, g, me, card, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two creatures were not sacrificed")
	}
	// The cast card left the hand again; the spell drew three.
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
}

func TestOldB176ClearTheMindShufflesTheTargetsGraveyardAndYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card1 := effectstest.PushGraveyardCardTyped(opp, "Dead A", "Sorcery")
	card2 := effectstest.PushGraveyardCardTyped(opp, "Dead B", "Sorcery")
	effectstest.SeedLibrary(me, "mine")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Clear the Mind", "Sorcery", oldB176ClearMind, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 || !opp.Library.Contains(card1) || !opp.Library.Contains(card2) {
		t.Errorf("their graveyard %d, library has the cards: %v %v", opp.Graveyard.Size(), opp.Library.Contains(card1), opp.Library.Contains(card2))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (drew one)", me.Hand.Size(), hand+1)
	}
}

func TestOldB176DomrisNodorogFetchesDomriFromTheGraveyardOrDeclines(t *testing.T) {
	for _, mode := range []string{"fetch", "decline"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		domri := effectstest.PushGraveyardCardTyped(me, "Domri, City Smasher", "Legendary Planeswalker — Domri")
		effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.CastCatalogSpell(t, g, "Domri's Nodorog", "Creature — Beast", oldB176Nodorog, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if mode == "decline" {
			effectstest.AnswerMayChoice(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Contains(domri) {
				t.Error("declining still took Domri")
			}
			continue
		}
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		if c := effectstest.SearchChoiceFor(g, me.ID); c != nil && c.SearchZonesPending() {
			effectstest.AnswerSearchZones(t, g, me.ID, "graveyard")
		}
		effectstest.AnswerSearchByID(t, g, me.ID, domri)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(domri) {
			t.Error("Domri did not reach the hand")
		}
	}
}

func TestOldB176TomeOfTheGuildpactDrawsOnMulticoloredSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tome of the Guildpact", "Artifact — Book", oldB176Tome, false)
	effectstest.SeedLibrary(me, "a", "b", "c")
	hand := me.Hand.Size()
	if err := oldB176Cast(t, g, me, game.Card{Name: "Mono", TypeLine: "Sorcery", Colors: []string{"R"}}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after a mono-colored spell, want %d", me.Hand.Size(), hand)
	}
	if err := oldB176Cast(t, g, me, game.Card{Name: "Gold", TypeLine: "Sorcery", Colors: []string{"R", "G"}}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after a multicolored spell, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB176TomeOfTheGuildpactTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tome := effectstest.PushCatalogPermanent(g, me.ID, "Tome of the Guildpact", "Artifact — Book", oldB176Tome, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, tome, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.CountColor("U") != 1 {
		t.Error("no blue mana")
	}
}

func TestOldB176DovinsDismissalTopdecksATappedCreatureAndMayFetchDovin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	dovin := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Dovin, Architect of Law", TypeLine: "Legendary Planeswalker — Dovin"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Dovin's Dismissal", "Instant", oldB176Dismissal, effectstest.CardRefs(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Dovin's Dismissal", "Instant", oldB176Dismissal, effectstest.CardRefs(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Library.Contains(tapped) {
		t.Error("the tapped creature is not in its owner's library")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil && c.SearchZonesPending() {
		effectstest.AnswerSearchZones(t, g, me.ID, "library")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, dovin)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dovin) {
		t.Error("Dovin did not reach the hand")
	}
}

func TestOldB176DovinsAcuityGainsDrawsAndMayReturnOnAnInstantInYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "a", "b", "c")
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Dovin's Acuity", "Enchantment", oldB176Acuity, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d (want %d), hand %d (want %d)", me.Life, life+2, me.Hand.Size(), hand+1)
	}
	var onField uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.OracleID == oldB176Acuity {
			onField = c.InstanceID
		}
	}
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(onField) || !me.Hand.Contains(onField) {
		t.Error("Dovin's Acuity did not return to hand")
	}
}

func TestOldB176VannifarSacrificesAnotherCreatureAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	van := effectstest.PushCatalogPermanent(g, me.ID, "Prime Speaker Vannifar", "Legendary Creature — Elf Ooze Wizard", oldB176Vannifar, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, fodder)
		c.ManaCost = "{1}{G}"
	})
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"},
		game.Card{Name: "Four Drop", TypeLine: "Creature — Bear", ManaCost: "{3}{G}"})
	var three uuid.UUID
	for _, c := range me.Library.Cards {
		if c.Name == "Three Drop" {
			three = c.InstanceID
		}
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, van, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{van}}); err == nil {
		t.Fatal("Vannifar sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, van, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil {
		t.Fatal("no search prompt")
	} else if effectstest.SearchOptionNamed(g, c, "Four Drop") != uuid.Nil {
		t.Error("the search offered a creature of the wrong mana value")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, three)
	if !g.Battlefield.Contains(three) {
		t.Error("the three-drop did not enter the battlefield")
	}
}

func TestOldB176AmplifireDoublesTheRevealedCreaturesBaseStatsAndBottomsTheRun(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fire := effectstest.PushCatalogPermanent(g, me.ID, "Amplifire", "Creature — Elemental", oldB176Amplifire, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, fire)
		c.Power, c.Toughness = 1, 1
	})
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Big", TypeLine: "Creature — Beast", Power: 3, Toughness: 4},
		game.Card{Name: "Deep", TypeLine: "Creature — Beast", Power: 9, Toughness: 9})
	_ = ids
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := effectstest.EffectivePTOf(t, g, fire)
	if p != 6 || tough != 8 {
		t.Errorf("Amplifire is %d/%d, want 6/8", p, tough)
	}
	// Spell and Big went to the bottom; the creature card behind them stays on top.
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.Name != "Deep" {
		t.Errorf("top of library %q, want Deep", top.Name)
	}
}
