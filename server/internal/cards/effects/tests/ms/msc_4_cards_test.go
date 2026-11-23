package testsms

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_4_cards_test.go — Marvel Super Heroes Commander, fourth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	msc4BovineOracle       = "7402d98f-6d58-4ff2-a0db-9eec4417a75f"
	msc4BlackTomOracle     = "c0d20e66-deb6-4a20-a2c1-8a79514e3792"
	msc4IntelExpertOracle  = "99c3a315-db53-4292-9d31-1efd791d3a13"
	msc4DeadlyHunterOracle = "12a7bf3c-5194-48ad-be0b-553b8ad67671"
	msc4DaringOpOracle     = "e8815686-8c47-4067-9dfa-744b2f5c1ba2"
	msc4AgileAvengerOracle = "df7abfb0-21df-4359-be81-0f3c09e1425e"
	msc4BlueMarvelOracle   = "16f9368a-5da5-4397-a127-c96cdd115b7b"
	msc4ClawsOfBastOracle  = "0a41221a-0cce-4aae-989a-24b70acca8db"
	msc4BobOracle          = "9caa490e-dc28-4069-9ffb-4b1afd8fdd54"
	msc4BuckyOracle        = "55bc5d06-7dd1-4403-af32-0e10d8bbe327"
	msc4UnbowedOracle      = "4fa2c9cb-106c-482c-9ba7-5bba2ea9e4a3"
	msc4SteveRogersOracle  = "d2084a6b-3248-4ed4-8fb5-d9c439d1ee72"
)

func TestMsc4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc4BovineOracle:       "Bovine Intervention",
		msc4BlackTomOracle:     "Black Tom Cassidy",
		msc4IntelExpertOracle:  "Black Widow, Intel Expert",
		msc4DeadlyHunterOracle: "Black Widow, Deadly Hunter",
		msc4DaringOpOracle:     "Black Widow, Daring Operative",
		msc4AgileAvengerOracle: "Black Widow, Agile Avenger",
		msc4BlueMarvelOracle:   "Blue Marvel, Adam Brashear",
		msc4ClawsOfBastOracle:  "Black Panther, Claws of Bast",
		msc4BobOracle:          "Bob, Reluctant HYDRA Agent",
		msc4BuckyOracle:        "Bucky Barnes, Eager Ally",
		msc4UnbowedOracle:      "Captain America, Unbowed",
		msc4SteveRogersOracle:  "Captain America, Steve Rogers",
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

// --- Bovine Intervention ---

func TestMsc4BovineInterventionDestroysACreatureAndGivesItsControllerAnOx(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bovine Intervention", "Instant", msc4BovineOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the targeted creature survived")
	}
	if n := effectstest.TokensNamedControlledBy(g, "Ox", opp.ID); n != 1 {
		t.Fatalf("%d Ox tokens for the creature's controller, want 1", n)
	}
	if n := effectstest.TokensNamedControlledBy(g, "Ox", g.Seats[0].ID); n != 0 {
		t.Errorf("the caster got %d Ox tokens, want 0", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ox" {
			if c.Power != 2 || c.Toughness != 2 || !c.HasColor("W") || !c.HasSubtype("Ox") || !c.IsCreature() {
				t.Errorf("the Ox is %d/%d %v %q, want a 2/2 white Ox creature", c.Power, c.Toughness, c.Colors, c.TypeLine)
			}
		}
	}
}

func TestMsc4BovineInterventionDestroysAnArtifactAndOwnTargetGetsTheOx(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Sol Ring Stand-in", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Bovine Intervention", "Instant", msc4BovineOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the targeted artifact survived")
	}
	if n := effectstest.TokensNamedControlledBy(g, "Ox", opp.ID); n != 1 {
		t.Errorf("%d Ox tokens for the artifact's controller, want 1", n)
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bovine Intervention", "Instant", msc4BovineOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Ox", me.ID); n != 1 {
		t.Errorf("the caster destroying their own creature got %d Ox tokens, want 1", n)
	}
}

func TestMsc4BovineInterventionCannotTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Bovine Intervention", "Instant", msc4BovineOracle, effectstest.B16TargetCard(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

// --- Black Tom Cassidy ---

func TestMsc4BlackTomAddsGreenAndGainsLifeOnlyWithAnotherMutant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tom := effectstest.B12Push(g, me.ID, "Black Tom Cassidy", "Legendary Creature — Mutant Villain", msc4BlackTomOracle, 1, 3)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, tom, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want [G]", got)
	}
	if me.Life != life {
		t.Errorf("life %d -> %d with no other Mutant, want no change", life, me.Life)
	}

	effectstest.B12Creature(g, me.ID, "Another Mutant", "Creature — Mutant", 2, 2)
	effectstest.Untap(g, tom)
	if err := g.ActivateManaAbility(me.ID, tom, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life+1 {
		t.Errorf("life %d -> %d with another Mutant, want +1", life, me.Life)
	}
}

func TestMsc4BlackTomIgnoresAnOpponentsMutant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tom := effectstest.B12Push(g, me.ID, "Black Tom Cassidy", "Legendary Creature — Mutant Villain", msc4BlackTomOracle, 1, 3)
	effectstest.B12Creature(g, opp.ID, "Their Mutant", "Creature — Mutant", 2, 2)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, tom, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life {
		t.Errorf("an opponent's Mutant gained life: %d -> %d", life, me.Life)
	}
}

// --- Black Widow, Intel Expert ---

func TestMsc4IntelExpertMakesBothPlayersDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := effectstest.B12Push(g, me.ID, "Black Widow, Intel Expert", "Legendary Creature — Human Spy Hero", msc4IntelExpertOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, widow, "first strike") {
		t.Error("Black Widow has no first strike")
	}
	mine, theirs, other := me.Hand.Size(), opp.Hand.Size(), g.Seats[2].Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, widow)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != mine+2 {
		t.Errorf("my hand %d -> %d, want +2", mine, got)
	}
	if got := opp.Hand.Size(); got != theirs+2 {
		t.Errorf("the damaged player's hand %d -> %d, want +2", theirs, got)
	}
	if got := g.Seats[2].Hand.Size(); got != other {
		t.Errorf("a bystander's hand %d -> %d, want no change", other, got)
	}
}

func TestMsc4IntelExpertDoesNothingWhenAnotherCreatureDealsTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Black Widow, Intel Expert", "Legendary Creature — Human Spy Hero", msc4IntelExpertOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != mine || opp.Hand.Size() != theirs {
		t.Errorf("hands %d/%d -> %d/%d after another creature connected, want no change", mine, theirs, me.Hand.Size(), opp.Hand.Size())
	}
}

// --- Black Widow, Deadly Hunter ---

func TestMsc4DeadlyHunterDrawsForEachDeathtouchCreatureThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.B12Push(g, me.ID, "Black Widow, Deadly Hunter", "Legendary Creature — Human Assassin Hero", msc4DeadlyHunterOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, hunter, bear)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1 (only the deathtouch creature counts)", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
}

func TestMsc4DeadlyHunterCountsAnotherCreatureWithDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Black Widow, Deadly Hunter", "Legendary Creature — Human Assassin Hero", msc4DeadlyHunterOracle, 3, 3)
	snake := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Deathtouch Snake", TypeLine: "Creature — Snake",
		Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, snake)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1 for another deathtouch creature", hand, got)
	}
}

// --- Black Widow, Daring Operative ---

func TestMsc4DaringOperativeMillsThreeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := me.Library.Size()
	effectstest.B38CastCreature(t, g, "Black Widow, Daring Operative", "Legendary Creature — Human Spy Hero", msc4DaringOpOracle, 3, 4)
	if got := me.Library.Size(); got != lib-3 {
		t.Errorf("library %d -> %d, want 3 milled", lib, got)
	}
}

func TestMsc4DaringOperativeDrainsByTheCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, hit, other := g.Seats[0], g.Seats[1], g.Seats[2]
	widow := effectstest.B12Push(g, me.ID, "Black Widow, Daring Operative", "Legendary Creature — Human Spy Hero", msc4DaringOpOracle, 3, 4)
	effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.PushGraveyardPermanent(me, "Dead Wurm", "Creature — Wurm", "{3}{G}")
	effectstest.PushGraveyardPermanent(me, "Old Relic", "Artifact", "{2}")
	hitLife, otherLife, mine := hit.Life, other.Life, me.Life
	effectstest.DeclareAttack(t, g, hit.ID, widow)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if hit.Life != hitLife-3-2 {
		t.Errorf("the damaged player %d -> %d, want 3 combat damage and 2 lost", hitLife, hit.Life)
	}
	if other.Life != otherLife-2 {
		t.Errorf("another opponent %d -> %d, want 2 lost", otherLife, other.Life)
	}
	if me.Life != mine {
		t.Errorf("my life %d -> %d, want no change", mine, me.Life)
	}
}

// --- Black Widow, Agile Avenger ---

func TestMsc4AgileAvengerGrowsAndDrawsOnAnOpponentsSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := effectstest.PushCatalogPermanent(g, me.ID, "Black Widow, Agile Avenger", "Legendary Creature — Human Spy Hero", msc4AgileAvengerOracle, false)
	hand := me.Hand.Size()
	effectstest.HobA5Draw(t, g, opp, nil)
	if n := effectstest.HobA5Counters(t, g, widow); n != 0 || me.Hand.Size() != hand {
		t.Fatalf("after one opposing draw: %d counters, hand %d -> %d, want none", n, hand, me.Hand.Size())
	}
	effectstest.HobA5Draw(t, g, opp, nil)
	if n := effectstest.HobA5Counters(t, g, widow); n != 1 {
		t.Errorf("%d counters after the second opposing draw, want 1", n)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
	effectstest.HobA5Draw(t, g, opp, nil)
	if n := effectstest.HobA5Counters(t, g, widow); n != 1 {
		t.Errorf("%d counters after the third opposing draw, want it to stay 1", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.HobA5Counters(t, g, widow); n != 1 {
		t.Errorf("%d counters after my own two draws, want it to stay 1", n)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, widow, "menace") {
		t.Error("Black Widow has no menace")
	}
}

// --- Blue Marvel, Adam Brashear ---

func TestMsc4BlueMarvelGrowsOnTheSecondDrawAndHasFlying(t *testing.T) {
	id := effectstest.HobA5SecondDrawGrows(t, "Blue Marvel, Adam Brashear", "Legendary Creature — Human Scientist Hero", msc4BlueMarvelOracle)
	_ = id
	g := effectstest.NewCatalogGame(t)
	marvel := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Blue Marvel, Adam Brashear", "Legendary Creature — Human Scientist Hero", msc4BlueMarvelOracle, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, marvel, "flying") {
		t.Error("Blue Marvel has no flying")
	}
}

// --- Black Panther, Claws of Bast ---

func TestMsc4ClawsOfBastCountersACreatureThatAttacksAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	panther := effectstest.B12Push(g, me.ID, "Black Panther, Claws of Bast", "Legendary Creature — Human Warrior Hero", msc4ClawsOfBastOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, panther, "lifelink") {
		t.Error("Black Panther has no lifelink")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, bear); n != 1 {
		t.Errorf("the lone attacker has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.PlusOnes(g, panther); n != 0 {
		t.Errorf("Black Panther, who stayed home, has %d counters, want 0", n)
	}
}

func TestMsc4ClawsOfBastDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Black Panther, Claws of Bast", "Legendary Creature — Human Warrior Hero", msc4ClawsOfBastOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, a) != 0 || effectstest.PlusOnes(g, b) != 0 {
		t.Error("a creature that did not attack alone got a counter")
	}
}

// --- Bob, Reluctant HYDRA Agent ---

func TestMsc4BobAttackingAloneReturnsToHandAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bob := effectstest.B12Push(g, me.ID, "Bob, Reluctant HYDRA Agent", "Legendary Creature — Human Coward Villain", msc4BobOracle, 2, 2)
	lives := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	mine := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, bob)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bob) || !effectstest.InHand(me, bob) {
		t.Error("Bob did not return to his owner's hand")
	}
	for i, before := range lives {
		if got := g.Seats[i+1].Life; got != before-2 {
			t.Errorf("opponent %d life %d -> %d, want -2", i+1, before, got)
		}
	}
	if me.Life != mine+2 {
		t.Errorf("my life %d -> %d, want +2", mine, me.Life)
	}
}

func TestMsc4BobAttackingWithAnotherCreatureStaysAndDrainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bob := effectstest.B12Push(g, me.ID, "Bob, Reluctant HYDRA Agent", "Legendary Creature — Human Coward Villain", msc4BobOracle, 0, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 0, 2)
	mine, theirs := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, bob, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bob) {
		t.Error("Bob returned to hand although he did not attack alone")
	}
	if me.Life != mine || opp.Life != theirs {
		t.Errorf("life %d/%d -> %d/%d, want no change", mine, theirs, me.Life, opp.Life)
	}
}

// --- Bucky Barnes, Eager Ally ---

func TestMsc4BuckyDyingLooksAtFourAndTakesAnEquipmentHeroOrSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bucky := effectstest.B12Push(g, me.ID, "Bucky Barnes, Eager Ally", "Legendary Creature — Human Soldier Hero", msc4BuckyOracle, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, bucky, "vigilance") {
		t.Error("Bucky has no vigilance")
	}
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Vibranium Shield", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Spare Soldier", TypeLine: "Creature — Human Soldier"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	lib := me.Library.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bucky) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the Equipment did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card taken and the rest put back", lib, me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went under it")
	}
}

// --- Captain America, Unbowed ---

func TestMsc4UnbowedGivesSoldiersAndHeroesYouControlIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	soldier := effectstest.B12Creature(g, me.ID, "Mine Soldier", "Creature — Human Soldier", 2, 2)
	hero := effectstest.B12Creature(g, me.ID, "Mine Hero", "Creature — Human Hero", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Mine Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 2, 2)
	cap := effectstest.B38CastCreature(t, g, "Captain America, Unbowed", "Legendary Creature — Human Soldier Hero", msc4UnbowedOracle, 3, 4)
	for name, id := range map[string]uuid.UUID{"Soldier": soldier, "Hero": hero, "Captain America": cap} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("your %s did not gain indestructible", name)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a Bear gained indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's Soldier gained indestructible")
	}
	spec, _ := effects.Lookup(msc4UnbowedOracle)
	if !reflect.DeepEqual(spec.PrintedKeywords, []string{"flash"}) {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Captain America, Steve Rogers ---

func TestMsc4SteveRogersCountersAndProtectsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steve := effectstest.B12Push(g, me.ID, "Captain America, Steve Rogers", "Legendary Creature — Human Soldier Hero", msc4SteveRogersOracle, 4, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, steve, "lifelink") {
		t.Error("Captain America has no lifelink")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 0, 2)
	effectstest.DeclareAttack(t, g, opp.ID, steve)
	msc4PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, bear); n != 1 {
		t.Errorf("the other creature has %d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the other creature did not gain indestructible")
	}
	if n := effectstest.PlusOnes(g, steve); n != 0 {
		t.Errorf("Captain America put a counter on himself: %d", n)
	}
}

// msc4PickIfAsked answers the trigger's target prompt with `target`
// when the engine opened one, and does nothing when it chose for us.
func msc4PickIfAsked(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if effectstest.LatestPickTarget(g, chooser) == nil {
		return
	}
	effectstest.PickCard(t, g, chooser, target)
}
