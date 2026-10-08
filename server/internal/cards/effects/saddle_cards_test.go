package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saddle_cards_test.go — one test per Mount (and Guidelight Matrix)
// that #2695 ships, each driving the whole path: saddle in the main
// phase, attack, resolve, and read the result off the board.

const (
	boundingFelidarOracle  = "b00f316d-7c82-40e6-80e5-9f36434590a3"
	gloryheathLynxOracle   = "28673657-f05f-4b0a-b80c-939773abdefb"
	seraphicSteedOracle    = "51622e89-aa5e-4dd6-a078-eb45e8ebb434"
	districtMascotOracle   = "5fa7cf3a-2a4d-4dd7-890b-49c497570b16"
	bulwarkOxOracle        = "bcd56fbf-fde3-41c2-83f5-9636e92d2893"
	guardianSunmareOracle  = "3b319540-4db1-43d3-8f4d-f5a276ad78e4"
	ornheryTumblewaggOrcl  = "f8d7bcbd-9a1d-4fc6-abdd-88a7e01b9411"
	causticBroncoOracle    = "165bdc68-1da0-47db-a069-978fc4c04b3f"
	saddleCardsLibraryLand = "Basic Land — Plains"
)

// saddleThenAttack saddles `mount` with `saddlers` in the first main
// phase, resolves the ability, and attacks `defender` with the Mount,
// leaving its attack trigger on the stack.
func saddleThenAttack(t *testing.T, g *game.Game, owner, defender, mount uuid.UUID, saddlers ...uuid.UUID) {
	t.Helper()
	toMain(t, g)
	if err := saddleAbility(t, g, owner, mount, saddlers...); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !saddleCard(t, g, mount).Saddled {
		t.Fatal("setup: the Mount is not saddled")
	}
	declareAttack(t, g, defender, mount)
}

func pushLibraryCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

func countersOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return saddleCard(t, g, id).Counters[game.CounterPlusOne]
}

func TestDroverGrizzlyGrantsTrampleToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Drover Grizzly", droverGrizzlyOracle, 3)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	bystander := pushCrewerForTest(g, me.ID, "Bystander", 1)
	foe := pushCrewerForTest(g, opp.ID, "Foe", 1)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mount, saddler, bystander} {
		if !effectiveAbilitiesContain(t, g, id, "trample") {
			t.Errorf("%s did not gain trample", saddleCard(t, g, id).Name)
		}
	}
	if effectiveAbilitiesContain(t, g, foe, "trample") {
		t.Error("an opponent's creature gained trample")
	}
}

func TestBoundingFelidarCountersEveryOtherCreatureAndGainsLifeForEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Bounding Felidar", boundingFelidarOracle, 4)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 2)
	bystander := pushCrewerForTest(g, me.ID, "Bystander", 1)
	foe := pushCrewerForTest(g, opp.ID, "Foe", 1)
	life := me.Life

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)

	if got := countersOf(t, g, saddler); got != 1 {
		t.Errorf("the saddler has %d counters, want 1", got)
	}
	if got := countersOf(t, g, bystander); got != 1 {
		t.Errorf("the bystander has %d counters, want 1", got)
	}
	if got := countersOf(t, g, mount); got != 0 {
		t.Errorf("the Felidar put a counter on itself (%d)", got)
	}
	if got := countersOf(t, g, foe); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (one per creature that got a counter)", got)
	}
}

func TestGloryheathLynxFetchesABasicPlainsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Gloryheath Lynx", gloryheathLynxOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 2)
	plains := pushLibraryCard(me, game.Card{Name: "Plains", TypeLine: saddleCardsLibraryLand})
	pushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	hand := me.Hand.Size()

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
	if !me.Hand.Contains(plains) {
		t.Error("the Plains was not the card fetched")
	}
}

func TestSeraphicSteedMakesAnAngelWhenItAttacksSaddled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Seraphic Steed", seraphicSteedOracle, 3)
	saddler := pushCrewerForTest(g, me.ID, "Big Saddler", 4)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)

	var angel *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Angel" {
			angel = &g.Battlefield.Cards[i]
		}
	}
	if angel == nil {
		t.Fatal("no Angel token")
	}
	if angel.CurrentPower() != 3 || angel.CurrentToughness() != 3 || !effectiveAbilitiesContain(t, g, angel.InstanceID, "flying") {
		t.Errorf("Angel = %d/%d, want a 3/3 with flying", angel.CurrentPower(), angel.CurrentToughness())
	}
}

func TestDistrictMascotEntersWithACounterGrowsAndDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "District Mascot", "Creature — Dog Mount", districtMascotOracle, nil)
	passPriorityAroundTable(t, g)
	var cast uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == districtMascotOracle {
			cast = c.InstanceID
		}
	}
	if cast == uuid.Nil {
		t.Fatal("the Mascot did not resolve")
	}
	if got := countersOf(t, g, cast); got != 1 {
		t.Fatalf("enters with %d counters, want 1", got)
	}

	// A Mascot that has been here a turn (this one is summoning sick)
	// attacks saddled and adds one more.
	mount := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "District Mascot", OracleID: districtMascotOracle,
		TypeLine: "Creature — Dog Mount", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)
	if got := countersOf(t, g, mount); got != 2 {
		t.Fatalf("after the saddled attack: %d counters, want 2", got)
	}

	// Two counters and {1}{G} destroy an artifact.
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	mkAdd(t, g, me, "{G}{G}", game.AddManaOptions{})
	if err := g.ActivateCatalogAbility(me.ID, mount, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCardByID(g, rock); ok {
		t.Error("the artifact was not destroyed")
	}
	if got := countersOf(t, g, mount); got != 0 {
		t.Errorf("%d counters left, want 0 (two were the cost)", got)
	}
}

func TestBulwarkOxPutsACounterOnTargetAndProtectsCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Bulwark Ox", bulwarkOxOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	plain := pushCrewerForTest(g, me.ID, "Plain", 1)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	pickTriggerTarget(t, g, me.ID, plain)
	passPriorityAroundTable(t, g)
	if got := countersOf(t, g, plain); got != 1 {
		t.Fatalf("the target got %d counters, want 1", got)
	}

	// Sacrifice: only the creature with a counter gains the keywords.
	other := pushCrewerForTest(g, me.ID, "Other", 1)
	if err := g.ActivateCatalogAbility(me.ID, mount, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectiveAbilitiesContain(t, g, plain, kw) {
			t.Errorf("the creature with a counter did not gain %s", kw)
		}
		if effectiveAbilitiesContain(t, g, other, kw) {
			t.Errorf("a creature with no counter gained %s", kw)
		}
	}
}

func TestGuardianSunmareFetchesANonlandPermanentOfManaValueThreeOrLess(t *testing.T) {
	spec, ok := Lookup(guardianSunmareOracle)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("Guardian Sunmare: want one declared Aura caveat, got %+v", spec.Caveats)
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Guardian Sunmare", guardianSunmareOracle, 4)
	saddler := pushCrewerForTest(g, me.ID, "Big Saddler", 4)
	tooBig := pushLibraryCard(me, game.Card{Name: "Too Big", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5})
	bolt := pushLibraryCard(me, game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	forest := pushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	target := pushLibraryCard(me, game.Card{Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 1, Toughness: 1})

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)

	var choice *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
			choice = c
		}
	}
	if choice == nil {
		t.Fatal("no search prompt")
	}
	// The legal picks are exactly the nonland permanent cards of mana value 3 or less.
	for _, id := range []uuid.UUID{tooBig, bolt, forest} {
		if err := g.ResolveSearchLibrary(choice.ID, me.ID, []uuid.UUID{id}); err == nil {
			t.Fatalf("an illegal card (%s) was accepted", id)
		}
	}
	if err := g.ResolveSearchLibrary(choice.ID, me.ID, []uuid.UUID{target}); err != nil {
		t.Fatalf("pick the Elf: %v", err)
	}
	if _, ok := battlefieldCardByID(g, target); !ok {
		t.Error("the chosen card was not put onto the battlefield")
	}
}

func TestOrneryTumblewaggCountersAtCombatAndDoublesOnASaddledAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Ornery Tumblewagg", ornheryTumblewaggOrcl, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 2)
	pump := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pump Target", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 3},
	})

	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Beginning of combat: a counter on the creature with three.
	advanceTo(t, g, game.StepBeginCombat)
	pickTriggerTarget(t, g, me.ID, pump)
	passPriorityAroundTable(t, g)
	if got := countersOf(t, g, pump); got != 4 {
		t.Fatalf("after the combat trigger: %d counters, want 4", got)
	}

	// The saddled attack doubles them.
	declareAttack(t, g, opp.ID, mount)
	pickTriggerTarget(t, g, me.ID, pump)
	passPriorityAroundTable(t, g)
	if got := countersOf(t, g, pump); got != 8 {
		t.Errorf("after the doubling: %d counters, want 8", got)
	}
}

func TestCausticBroncoPaysByWhetherItIsSaddled(t *testing.T) {
	for _, tc := range []struct {
		name    string
		saddled bool
	}{{"unsaddled: you lose the life", false}, {"saddled: each opponent loses it", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mount := pushMountForTest(g, me.ID, "Caustic Bronco", causticBroncoOracle, 3)
			top := pushLibraryCard(me, game.Card{Name: "Three Drop", TypeLine: "Creature — Test", ManaCost: "{2}{B}"})
			myLife, theirLife := me.Life, opp.Life
			hand := me.Hand.Size()

			if tc.saddled {
				saddler := pushCrewerForTest(g, me.ID, "Big Saddler", 3)
				saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
			} else {
				declareAttack(t, g, opp.ID, mount)
			}
			passPriorityAroundTable(t, g)

			if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
				t.Error("the revealed card did not go to hand")
			}
			if tc.saddled {
				if me.Life != myLife || opp.Life != theirLife-3 {
					t.Errorf("life me %d→%d, opp %d→%d; want only the opponent to lose 3", myLife, me.Life, theirLife, opp.Life)
				}
			} else if me.Life != myLife-3 || opp.Life != theirLife {
				t.Errorf("life me %d→%d, opp %d→%d; want only you to lose 3", myLife, me.Life, theirLife, opp.Life)
			}
		})
	}
}

func TestStubbornBurrowfiendGetsPlusXPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Stubborn Burrowfiend", stubbornBurrowfiendOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 2)
	for i := 0; i < 2; i++ {
		pushLibraryCard(me, game.Card{Name: "Milled Creature", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	}
	pushGraveyardPermanent(me, "Old Creature", "Creature — Test", "{1}")

	if err := saddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Three creature cards in the graveyard after the mill: +3/+3 on a 2/2.
	if got := saddleCard(t, g, mount).CurrentPower(); got != 5 {
		t.Errorf("power = %d, want 5 (2 + one per creature card in the graveyard)", got)
	}
}
