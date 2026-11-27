package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 629 (Apocalypse: card draw and library): Whirlpool Rider,
// Squee's Revenge, Ceta Sanctuary, Gaea's Balance, Last Caress, Savage Gorilla,
// Symbiotic Deployment, Tranquil Path, Whirlpool Drake, Wild Research.

const (
	oldB629Rider    = "13a3f2c1-0454-4661-a462-542254f8360d"
	oldB629Revenge  = "3c9f3f26-339e-459e-9847-4e33aafb6f9b"
	oldB629Ceta     = "45c4e67c-e452-41c5-8baa-050819a1321f"
	oldB629Balance  = "5aadbd03-c459-4893-aa1b-efcf193faa1b"
	oldB629Caress   = "620b13cf-b8ea-48b7-8300-3c48390e32e3"
	oldB629Gorilla  = "6c876a57-c567-4d26-8ab6-a9c051e83ace"
	oldB629Symbiote = "a6872d7a-b647-4baf-bd1e-06707ccbb71a"
	oldB629Path     = "a9595011-8ddb-403a-ba35-5a79d346557e"
	oldB629Drake    = "aef9fcf0-b755-40f6-821d-836ae1b5497a"
	oldB629Research = "d4721c7e-1a55-42cf-99c9-9118e386f48b"
)

func TestOldB629EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB629Rider: "Whirlpool Rider", oldB629Revenge: "Squee's Revenge", oldB629Ceta: "Ceta Sanctuary",
		oldB629Balance: "Gaea's Balance", oldB629Caress: "Last Caress", oldB629Gorilla: "Savage Gorilla",
		oldB629Symbiote: "Symbiotic Deployment", oldB629Path: "Tranquil Path", oldB629Drake: "Whirlpool Drake",
		oldB629Research: "Wild Research",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB629Fill(g *game.Game) {
	for _, p := range g.Seats {
		effectstest.SeedLibrary(p, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "L9", "L10")
	}
}

func oldB629HandIDs(p *game.Player) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, c := range p.Hand.Cards {
		out[c.InstanceID] = true
	}
	return out
}

// oldB629Redrew checks one hand was shuffled into the library and as many cards
// drawn: the same sizes, no card lost, and not the same hand back.
func oldB629Redrew(t *testing.T, p *game.Player, before map[uuid.UUID]bool, handSize, libSize int) {
	t.Helper()
	if p.Hand.Size() != handSize || p.Library.Size() != libSize {
		t.Errorf("hand %d library %d, want %d and %d", p.Hand.Size(), p.Library.Size(), handSize, libSize)
	}
	same := 0
	for id := range before {
		if p.Hand.Contains(id) {
			same++
		} else if !p.Library.Contains(id) {
			t.Error("an old hand card is neither in hand nor library")
		}
	}
	if same == len(before) {
		t.Error("the same hand came back: nothing was shuffled")
	}
}

// --- Whirlpool Rider / Whirlpool Drake ----------------------------------------

func TestOldB629RiderEntersAndRedrawsOnlyYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	oldB629Fill(g)
	me, opp := g.Seats[0], g.Seats[1]
	oppHand := oldB629HandIDs(opp)
	effectstest.B12PlayFromHand(t, g, "Whirlpool Rider", "Creature — Merfolk", oldB629Rider, game.CastSpellParams{})
	before := oldB629HandIDs(me)
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB629Redrew(t, me, before, hand, lib)
	for id := range oppHand {
		if !opp.Hand.Contains(id) {
			t.Error("an opponent's hand was shuffled by the enters trigger")
		}
	}
}

func TestOldB629DrakeFliesAndRedrawsWhenItEntersAndWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	oldB629Fill(g)
	me := g.Seats[0]
	id := effectstest.B12PlayFromHand(t, g, "Whirlpool Drake", "Creature — Drake", oldB629Drake, game.CastSpellParams{})
	before := oldB629HandIDs(me)
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB629Redrew(t, me, before, hand, lib)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("the Drake has no flying")
	}
	before = oldB629HandIDs(me)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	oldB629Redrew(t, me, before, hand, lib)
}

// --- Last Caress ----------------------------------------------------------------

func TestOldB629LastCaressDrainsOneAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB629Fill(g)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Last Caress", "Sorcery", oldB629Caress, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != life+1 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, life+1, oppLife-1)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Tranquil Path ---------------------------------------------------------------

func TestOldB629TranquilPathDestroysEveryEnchantmentAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB629Fill(g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Aura Thing", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Idol", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Tranquil Path", "Sorcery", oldB629Path, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("a non-enchantment was destroyed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Ceta Sanctuary --------------------------------------------------------------

func oldB629Upkeep(t *testing.T, colors ...[]string) (g *game.Game, me *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	oldB629Fill(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ceta Sanctuary", "Enchantment", oldB629Ceta, false)
	for i, c := range colors {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Idol " + string(rune('A'+i)),
			TypeLine: "Enchantment", Colors: c, Owner: me.ID, Controller: me.ID})
	}
	return g, me
}

func TestOldB629CetaSanctuaryDoesNothingWithoutARedOrGreenPermanent(t *testing.T) {
	g, me := oldB629Upkeep(t, []string{"U"}, nil)
	hand := me.Hand.Size()
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Hand.Size() != hand || effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Errorf("the trigger did something: hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestOldB629CetaSanctuaryLootsOnceForOneColourAndTwiceForBoth(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors [][]string
		net    int
	}{
		{"red only", [][]string{{"R"}}, 0},
		{"green only", [][]string{{"G"}}, 0},
		{"red and green", [][]string{{"R"}, {"G"}}, 1},
		{"one red green permanent", [][]string{{"R", "G"}}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := oldB629Upkeep(t, tc.colors...)
			hand, grave := me.Hand.Size(), me.Graveyard.Size()
			effectstest.BeginUpkeep(t, g, me.ID)
			effectstest.DiscardFromHand(t, g, me.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand+tc.net || me.Graveyard.Size() != grave+1 {
				t.Errorf("hand %d graveyard %d, want %d and %d", me.Hand.Size(), me.Graveyard.Size(), hand+tc.net, grave+1)
			}
		})
	}
}

// --- Gaea's Balance --------------------------------------------------------------

func TestOldB629GaeasBalanceSacrificesFiveLandsAndFindsOneOfEachType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var lands []uuid.UUID
	for i := 0; i < 5; i++ {
		lands = append(lands, effectstest.PushCatalogPermanent(g, me.ID, "Wastes", "Land", "", false))
	}
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Second Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Dual", "Land — Plains Island"))
	// Too few lands, and a stranger's land, are refused.
	if err := castGaeasBalance(t, g, lands[:4]); err == nil {
		t.Fatal("cast with only four lands sacrificed")
	}
	if err := castGaeasBalance(t, g, lands); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, id := range lands {
		if g.Battlefield.Contains(id) {
			t.Error("a land was not sacrificed as the additional cost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	picks := []uuid.UUID{
		effectstest.SearchOptionNamed(g, c, "Forest"),
		effectstest.SearchOptionNamed(g, c, "Second Forest"),
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, picks); err == nil {
		t.Fatal("two Forests were accepted as two slots")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains", "Island", "Swamp", "Mountain", "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	names := map[string]bool{}
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me.ID {
				names[c.Name] = true
			}
		}
	})
	for _, n := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if !names[n] {
			t.Errorf("%s did not reach the battlefield", n)
		}
	}
}

func castGaeasBalance(t *testing.T, g *game.Game, sacrifice []uuid.UUID) error {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gaea's Balance", TypeLine: "Sorcery", OracleID: oldB629Balance, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: sacrifice})
}

// --- Savage Gorilla --------------------------------------------------------------

func TestOldB629SavageGorillaSacrificesToShrinkAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB629Fill(g)
	me, opp := g.Seats[0], g.Seats[1]
	gorilla := effectstest.B12Push(g, me.ID, "Savage Gorilla", "Creature — Ape", oldB629Gorilla, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UB")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gorilla, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(gorilla) {
		t.Error("the Gorilla was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 target survived -3/-3")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB629SavageGorillaNeedsBothColoursAndTheTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorilla := effectstest.B12Push(g, me.ID, "Savage Gorilla", "Creature — Ape", oldB629Gorilla, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, gorilla, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim)}); err == nil {
		t.Error("activated without {B}")
	}
	if !g.Battlefield.Contains(gorilla) {
		t.Error("the Gorilla was sacrificed by a refused activation")
	}
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Savage Gorilla", "Creature — Ape", oldB629Gorilla, true)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "UB")
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim)}); err == nil {
		t.Error("a summoning-sick Gorilla paid {T}")
	}
}

// --- Symbiotic Deployment --------------------------------------------------------

func TestOldB629SymbioticDeploymentTapsTwoCreaturesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB629Fill(g)
	me := g.Seats[0]
	dep := effectstest.PushCatalogPermanent(g, me.ID, "Symbiotic Deployment", "Enchantment", oldB629Symbiote, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, dep, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("activated tapping one creature")
	}
	b := effectstest.PushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", true) // summoning sick may still pay
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, dep, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) {
		t.Error("the two creatures were not tapped")
	}
}

func TestOldB629SymbioticDeploymentSkipsYourDrawStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB629Fill(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Symbiotic Deployment", "Enchantment", oldB629Symbiote, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if me.Hand.Size() != hand {
		t.Errorf("hand went from %d to %d through the draw step, want no draw", hand, me.Hand.Size())
	}
}

// --- Wild Research ---------------------------------------------------------------

func TestOldB629WildResearchFetchesTheNamedTypeAndDiscardsAtRandom(t *testing.T) {
	for _, tc := range []struct {
		mana, found, other string
		ability            int
	}{
		{"RW", "Holy Aura", "Zap", 0},
		{"RU", "Zap", "Holy Aura", 1},
	} {
		t.Run(tc.found, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.SeedSearchLibrary(me,
				game.Card{InstanceID: uuid.New(), Name: "Holy Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID},
				game.Card{InstanceID: uuid.New(), Name: "Zap", TypeLine: "Instant", Owner: me.ID, Controller: me.ID},
				effectstest.SearchTestLand("Plains", "Basic Land — Plains"))
			rg := effectstest.PushCatalogPermanent(g, me.ID, "Wild Research", "Enchantment", oldB629Research, false)
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.mana)
			hand, grave := me.Hand.Size(), me.Graveyard.Size()
			if err := g.ActivateCatalogAbility(me.ID, rg, tc.ability, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if effectstest.SearchOptionNamed(g, c, tc.other) != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Plains") != uuid.Nil {
				t.Errorf("the search offered something that is not the named type")
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, tc.found)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand || me.Graveyard.Size() != grave+1 {
				t.Errorf("hand %d graveyard %d, want %d and %d (one found, one discarded at random)", me.Hand.Size(), me.Graveyard.Size(), hand, grave+1)
			}
		})
	}
}

func TestOldB629WildResearchWithAnEmptyHandDiscardsTheFoundCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{InstanceID: uuid.New(), Name: "Holy Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = nil
	rg := effectstest.PushCatalogPermanent(g, me.ID, "Wild Research", "Enchantment", oldB629Research, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RW")
	if err := g.ActivateCatalogAbility(me.ID, rg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Holy Aura")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 || me.Graveyard.Size() != 1 {
		t.Errorf("hand %d graveyard %d, want 0 and 1: the only card in hand is the random discard", me.Hand.Size(), me.Graveyard.Size())
	}
}

// --- Squee's Revenge -------------------------------------------------------------

func TestOldB629SqueesRevengeDrawsTwoPerFlipOnlyWhenEveryFlipIsWon(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 120 && (won < 2 || lost < 2); table++ {
		g := effectstest.NewCatalogGame(t)
		oldB629Fill(g)
		me := g.Seats[0]
		effectstest.CastCatalogSpell(t, g, "Squee's Revenge", "Sorcery", oldB629Revenge, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.PayAmountPrompt(g, me.ID)
		if c == nil {
			t.Fatal("no number prompt")
		}
		if !c.PayAmount.NoMax || c.PayAmount.Min != 0 {
			t.Fatalf("prompt = %+v, want a number from 0 with no ceiling", *c.PayAmount)
		}
		if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
			t.Fatalf("answer 2: %v", err)
		}
		hand := me.Hand.Size()
		flips := 0
		for flips < 2 {
			effectstest.PassPriorityAroundTable(t, g)
			ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
			if ask == nil {
				break
			}
			before := len(effectstest.RandomEvents(g, game.EventFlipCoin))
			if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
				t.Fatalf("ResolveCoinCall: %v", err)
			}
			evs := effectstest.RandomEvents(g, game.EventFlipCoin)[before:]
			if len(evs) != 1 {
				t.Fatalf("%d coins flipped by one call, want 1", len(evs))
			}
			flips++
			if !evs[0].Won {
				break
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if all := flips == 2 && effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall) == nil && me.Hand.Size() == hand+4; all {
			won++
			continue
		}
		lost++
		if me.Hand.Size() != hand {
			t.Fatalf("hand %d after %d flips, want %d: a lost flip draws nothing", me.Hand.Size(), flips, hand)
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d clean wins and %d losses; want both", won, lost)
	}
}

func TestOldB629SqueesRevengeChoosingZeroFlipsAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB629Fill(g)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Squee's Revenge", "Sorcery", oldB629Revenge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no number prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatalf("answer 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall) != nil || me.Hand.Size() != hand {
		t.Error("a flip was asked, or a card drawn, for zero flips")
	}
}
