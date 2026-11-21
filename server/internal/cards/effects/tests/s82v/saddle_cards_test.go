package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDroverGrizzlyGrantsTrampleToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Drover Grizzly", effectstest.DroverGrizzlyOracle, 3)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)
	bystander := effectstest.PushCrewerForTest(g, me.ID, "Bystander", 1)
	foe := effectstest.PushCrewerForTest(g, opp.ID, "Foe", 1)

	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mount, saddler, bystander} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "trample") {
			t.Errorf("%s did not gain trample", effectstest.SaddleCard(t, g, id).Name)
		}
	}
	if effectstest.EffectiveAbilitiesContain(t, g, foe, "trample") {
		t.Error("an opponent's creature gained trample")
	}
}

func TestBoundingFelidarCountersEveryOtherCreatureAndGainsLifeForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Bounding Felidar", effectstest.BoundingFelidarOracle, 4)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 2)
	bystander := effectstest.PushCrewerForTest(g, me.ID, "Bystander", 1)
	foe := effectstest.PushCrewerForTest(g, opp.ID, "Foe", 1)
	life := me.Life

	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOf(t, g, saddler); got != 1 {
		t.Errorf("the saddler has %d counters, want 1", got)
	}
	if got := effectstest.CountersOf(t, g, bystander); got != 1 {
		t.Errorf("the bystander has %d counters, want 1", got)
	}
	if got := effectstest.CountersOf(t, g, mount); got != 0 {
		t.Errorf("the Felidar put a counter on itself (%d)", got)
	}
	if got := effectstest.CountersOf(t, g, foe); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (one per creature that got a counter)", got)
	}
}

func TestGloryheathLynxFetchesABasicPlainsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Gloryheath Lynx", effectstest.GloryheathLynxOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 2)
	plains := effectstest.PushLibraryCard(me, game.Card{Name: "Plains", TypeLine: effectstest.SaddleCardsLibraryLand})
	effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	hand := me.Hand.Size()

	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
	if !me.Hand.Contains(plains) {
		t.Error("the Plains was not the card fetched")
	}
}

func TestSeraphicSteedMakesAnAngelWhenItAttacksSaddled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Seraphic Steed", effectstest.SeraphicSteedOracle, 3)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Big Saddler", 4)

	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)

	var angel *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Angel" {
			angel = &g.Battlefield.Cards[i]
		}
	}
	if angel == nil {
		t.Fatal("no Angel token")
	}
	if angel.CurrentPower() != 3 || angel.CurrentToughness() != 3 || !effectstest.EffectiveAbilitiesContain(t, g, angel.InstanceID, "flying") {
		t.Errorf("Angel = %d/%d, want a 3/3 with flying", angel.CurrentPower(), angel.CurrentToughness())
	}
}

func TestDistrictMascotEntersWithACounterGrowsAndDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "District Mascot", "Creature — Dog Mount", effectstest.DistrictMascotOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var cast uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == effectstest.DistrictMascotOracle {
			cast = c.InstanceID
		}
	}
	if cast == uuid.Nil {
		t.Fatal("the Mascot did not resolve")
	}
	if got := effectstest.CountersOf(t, g, cast); got != 1 {
		t.Fatalf("enters with %d counters, want 1", got)
	}

	// A Mascot that has been here a turn (this one is summoning sick)
	// attacks saddled and adds one more.
	mount := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "District Mascot", OracleID: effectstest.DistrictMascotOracle,
		TypeLine: "Creature — Dog Mount", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)
	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOf(t, g, mount); got != 2 {
		t.Fatalf("after the saddled attack: %d counters, want 2", got)
	}

	// Two counters and {1}{G} destroy an artifact.
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.MkAdd(t, g, me, "{G}{G}", game.AddManaOptions{})
	if err := g.ActivateCatalogAbility(me.ID, mount, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCardByID(g, rock); ok {
		t.Error("the artifact was not destroyed")
	}
	if got := effectstest.CountersOf(t, g, mount); got != 0 {
		t.Errorf("%d counters left, want 0 (two were the cost)", got)
	}
}

func TestBulwarkOxPutsACounterOnTargetAndProtectsCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Bulwark Ox", effectstest.BulwarkOxOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)
	plain := effectstest.PushCrewerForTest(g, me.ID, "Plain", 1)

	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PickTriggerTarget(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOf(t, g, plain); got != 1 {
		t.Fatalf("the target got %d counters, want 1", got)
	}

	// Sacrifice: only the creature with a counter gains the keywords.
	other := effectstest.PushCrewerForTest(g, me.ID, "Other", 1)
	if err := g.ActivateCatalogAbility(me.ID, mount, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, plain, kw) {
			t.Errorf("the creature with a counter did not gain %s", kw)
		}
		if effectstest.EffectiveAbilitiesContain(t, g, other, kw) {
			t.Errorf("a creature with no counter gained %s", kw)
		}
	}
}

func TestGuardianSunmareFetchesANonlandPermanentOfManaValueThreeOrLess(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.GuardianSunmareOracle)
	// #3158: a fetched Aura chooses its host as it enters.
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Guardian Sunmare: want complete with no caveat, got %+v", spec.Caveats)
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Guardian Sunmare", effectstest.GuardianSunmareOracle, 4)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Big Saddler", 4)
	tooBig := effectstest.PushLibraryCard(me, game.Card{Name: "Too Big", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5})
	bolt := effectstest.PushLibraryCard(me, game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	forest := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	target := effectstest.PushLibraryCard(me, game.Card{Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 1, Toughness: 1})

	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)

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
	if _, ok := effectstest.BattlefieldCardByID(g, target); !ok {
		t.Error("the chosen card was not put onto the battlefield")
	}
}

func TestOrneryTumblewaggCountersAtCombatAndDoublesOnASaddledAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Ornery Tumblewagg", effectstest.OrnheryTumblewaggOrcl, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 2)
	pump := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pump Target", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 3},
	})

	effectstest.ToMain(t, g)
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Beginning of combat: a counter on the creature with three.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, pump)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOf(t, g, pump); got != 4 {
		t.Fatalf("after the combat trigger: %d counters, want 4", got)
	}

	// The saddled attack doubles them.
	effectstest.DeclareAttack(t, g, opp.ID, mount)
	effectstest.PickTriggerTarget(t, g, me.ID, pump)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOf(t, g, pump); got != 8 {
		t.Errorf("after the doubling: %d counters, want 8", got)
	}
}

func TestCausticBroncoPaysByWhetherItIsSaddled(t *testing.T) {
	for _, tc := range []struct {
		name    string
		saddled bool
	}{{"unsaddled: you lose the life", false}, {"saddled: each opponent loses it", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mount := effectstest.PushMountForTest(g, me.ID, "Caustic Bronco", effectstest.CausticBroncoOracle, 3)
			top := effectstest.PushLibraryCard(me, game.Card{Name: "Three Drop", TypeLine: "Creature — Test", ManaCost: "{2}{B}"})
			myLife, theirLife := me.Life, opp.Life
			hand := me.Hand.Size()

			if tc.saddled {
				saddler := effectstest.PushCrewerForTest(g, me.ID, "Big Saddler", 3)
				effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
			} else {
				effectstest.DeclareAttack(t, g, opp.ID, mount)
			}
			effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Stubborn Burrowfiend", effectstest.StubbornBurrowfiendOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 2)
	for i := 0; i < 2; i++ {
		effectstest.PushLibraryCard(me, game.Card{Name: "Milled Creature", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	}
	effectstest.PushGraveyardPermanent(me, "Old Creature", "Creature — Test", "{1}")

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Three creature cards in the graveyard after the mill: +3/+3 on a 2/2.
	if got := effectstest.SaddleCard(t, g, mount).CurrentPower(); got != 5 {
		t.Errorf("power = %d, want 5 (2 + one per creature card in the graveyard)", got)
	}
}
