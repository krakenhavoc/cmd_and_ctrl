package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_empower_a_test.go — tracker #2795, slice
// fra-empower-a: the twenty remaining "Empower Jace" cards, one test (or
// a few) each.

const (
	fraAvatarOracle        = "08e8a63c-9fcc-48cf-a9af-15cbaeec943f"
	fraFateholdCharmOracle = "313aee0e-4090-4589-b323-a4edbda21c68"
	fraHexhavenOracle      = "158823ed-6015-437a-82b0-8dadb1e9aab2"
	fraTethermageOracle    = "80c2784b-f27d-4a51-9380-171eb32b961d"
	fraMachinationsOracle  = "89bd056d-8f5b-4d0e-b80a-58a984e21100"
	fraOverwriteOracle     = "d0372cde-bf30-4a1b-93bb-9ff4ac1c116d"
	fraLurkerOracle        = "3ebd64d2-c178-45a2-a51f-9869754baa0f"
	fraProxyOracle         = "0089acfe-da66-4dd7-b1e5-4d7407f58257"
	fraSanctumOracle       = "37513bd8-7303-4ac6-a7a7-224184faf758"
	fraEchoesOracle        = "e8ecec61-cc47-4051-a599-0e7470764488"
	fraCryoOracle          = "7d32c63d-aa9e-4f5d-a3e6-51377a786009"
	fraDeathbringerOracle  = "54775625-bf44-4035-969e-7e70578c8b98"
	fraHealerOracle        = "cb27dc83-85ee-4f54-b032-41cd1806d3ac"
	fraMentorOracle        = "ce924285-f4c5-44e6-909b-f5f7d302ebaf"
	fraMindSculptorOracle  = "94be2e86-ba0e-4b8d-9b9c-114adacb26db"
	fraNecromancerOracle   = "f3087daa-d2ff-4227-abff-3398da6297af"
	fraParadoxOracle       = "76771cd1-6b74-49bb-9421-8845fe344567"
	fraPyromancerOracle    = "68d3f547-674a-4d8a-b02f-01bc2f161916"
	fraWarlordOracle       = "10721d62-dc68-43b4-bfa2-dcb8cf8ea980"
	fraWildspeakerOracle   = "501580b9-692c-4804-be34-012a25ba8adf"

	fraEnchantmentLine = "Legendary Enchantment"
)

// fraActivateGranted activates the granted loyalty row starting with
// `prefix` on a walker and settles the stack up to any prompt.
func fraActivateGranted(t *testing.T, g *game.Game, who, walker uuid.UUID, prefix string, params game.ActivateAbilityParams) {
	t.Helper()
	idx, ref := grantedLoyaltyRow(t, g, walker, prefix)
	params.Ref = ref
	if err := g.ActivateCatalogAbility(who, walker, idx, params); err != nil {
		t.Fatalf("activate %q: %v", prefix, err)
	}
	passPriorityAroundTable(t, g)
}

func fraCountNamed(g *game.Game, name string) int { return b16CountNamed(g, name) }

// castWay casts a Way, which enters and empowers Jace n.
func fraCastWay(t *testing.T, g *game.Game, name, oracle string) {
	t.Helper()
	castCatalogSpell(t, g, name, fraEnchantmentLine, oracle, nil)
	passPriorityAroundTable(t, g)
}

func TestWaysEnterAndEmpowerJace(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Way of the Cryomancer", fraCryoOracle, 5},
		{"Way of the Deathbringer", fraDeathbringerOracle, 5},
		{"Way of the Healer", fraHealerOracle, 5},
		{"Way of the Mentor", fraMentorOracle, 5},
		{"Way of the Mind Sculptor", fraMindSculptorOracle, 5},
		{"Way of the Necromancer", fraNecromancerOracle, 2},
		{"Way of the Paradox", fraParadoxOracle, 5},
		{"Way of the Pyromancer", fraPyromancerOracle, 2},
		{"Way of the Warlord", fraWarlordOracle, 5},
		{"Way of the Wildspeaker", fraWildspeakerOracle, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			fraCastWay(t, g, tc.name, tc.oracle)
			if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != tc.n {
				t.Fatalf("Jace loyalty %d, want %d", got, tc.n)
			}
		})
	}
}

func TestWayOfTheDeathbringerSacrificesForABeast(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Deathbringer", fraEnchantmentLine, fraDeathbringerOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	jace := seedJaceToken(t, g, me.ID, 4)

	fraActivateGranted(t, g, me.ID, jace, "−2", game.ActivateAbilityParams{})
	if got := loyaltyCount(g, jace); got != 2 {
		t.Fatalf("loyalty after −2: %d, want 2", got)
	}
	answerMayChoice(t, g, me.ID, true)
	answerSacrifice(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	beasts := battlefieldIDsNamed(g, "Beast")
	if len(beasts) != 1 {
		t.Fatalf("Beasts = %d, want 1", len(beasts))
	}
	if effectivePower(t, g, beasts[0]) != 4 || !effectiveAbilitiesContain(t, g, beasts[0], "trample") {
		t.Error("the Beast is not a 4/4 with trample")
	}
}

func TestWayOfTheDeathbringerDeclinedMakesNoBeast(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Deathbringer", fraEnchantmentLine, fraDeathbringerOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	jace := seedJaceToken(t, g, me.ID, 4)

	fraActivateGranted(t, g, me.ID, jace, "−2", game.ActivateAbilityParams{})
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !onBattlefieldNow(g, bear) || fraCountNamed(g, "Beast") != 0 {
		t.Error("declining the sacrifice still made a Beast or lost the creature")
	}
}

func TestWayOfTheWildspeakerMakesABeast(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Wildspeaker", fraEnchantmentLine, fraWildspeakerOracle, false)
	jace := seedJaceToken(t, g, me.ID, 4)
	short := seedJaceToken(t, g, me.ID, 3)

	idx, ref := grantedLoyaltyRow(t, g, short, "−4")
	if err := g.ActivateCatalogAbility(me.ID, short, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−4 with 3 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	fraActivateGranted(t, g, me.ID, jace, "−4", game.ActivateAbilityParams{})
	beasts := battlefieldIDsNamed(g, "Beast")
	if len(beasts) != 1 || effectivePower(t, g, beasts[0]) != 4 || !effectiveAbilitiesContain(t, g, beasts[0], "trample") {
		t.Fatalf("want one 4/4 trampling Beast, got %v", beasts)
	}
}

func TestWayOfTheHealerMakesACadetAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Healer", fraEnchantmentLine, fraHealerOracle, false)
	jace := seedJaceToken(t, g, me.ID, 4)

	fraActivateGranted(t, g, me.ID, jace, "−2", game.ActivateAbilityParams{})
	cadets := battlefieldIDsNamed(g, "Cadet")
	if len(cadets) != 1 {
		t.Fatalf("Cadets = %d, want 1", len(cadets))
	}
	c := cardByID(g, cadets[0])
	if c.TypeLine != "Token Creature — Wizard Soldier" || effectivePower(t, g, cadets[0]) != 2 || len(c.Colors) != 0 {
		t.Errorf("Cadet = %q %d power colours %v", c.TypeLine, effectivePower(t, g, cadets[0]), c.Colors)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil && latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("no surveil prompt was queued")
	}
}

func TestWayOfThePyromancerAddsRed(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Pyromancer", fraEnchantmentLine, fraPyromancerOracle, false)
	jace := seedJaceToken(t, g, me.ID, 1)

	fraActivateGranted(t, g, me.ID, jace, "+1", game.ActivateAbilityParams{})
	if got := loyaltyCount(g, jace); got != 2 {
		t.Errorf("loyalty after +1: %d, want 2", got)
	}
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool = %d, want one red", got)
	}
}

func TestWayOfTheWarlordHitsCreatureAndPlayer(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Way of the Warlord", fraEnchantmentLine, fraWarlordOracle, false)
	jace := seedJaceToken(t, g, me.ID, 5)
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life

	fraActivateGranted(t, g, me.ID, jace, "−4", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if onBattlefieldNow(g, bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

func TestWayOfTheWarlordNoCreatureStillHitsThePlayer(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Way of the Warlord", fraEnchantmentLine, fraWarlordOracle, false)
	jace := seedJaceToken(t, g, me.ID, 5)
	life := opp.Life
	fraActivateGranted(t, g, me.ID, jace, "−4", game.ActivateAbilityParams{
		// Slot 1 names the player clause; the optional creature clause is empty.
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}},
	})
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d (up to one creature: none chosen)", opp.Life, life-2)
	}
}

func TestWayOfTheCryomancerCopiesTheNextInstant(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Cryomancer", fraEnchantmentLine, fraCryoOracle, false)
	jace := seedJaceToken(t, g, me.ID, 4)
	fraActivateGranted(t, g, me.ID, jace, "−3", game.ActivateAbilityParams{})

	handBefore := handSize(me)
	// An instant that draws a card: the copy draws a second.
	castCatalogSpell(t, g, "Fixture Draw", "Instant", fraDrawOracle, nil)
	passPriorityAroundTable(t, g)
	// A copy trigger may offer new targets; none here.
	passPriorityAroundTable(t, g)
	// one card left the hand (the cast), two drawn
	if got := handSize(me); got != handBefore+1 && got != handBefore+2 {
		t.Fatalf("hand %d -> %d after a copied draw spell", handBefore, got)
	}
}

const fraDrawOracle = "test-fra-empower-a-draw"

func init() {
	Register(Spec{
		OracleID: fraDrawOracle,
		Name:     "Fixture Draw",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}

func TestWayOfTheMentorGrowsWalkersOnLifegain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Way of the Mentor", fraEnchantmentLine, fraMentorOracle, false)
	a := seedJaceToken(t, g, me.ID, 3)
	b := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)
	theirs := fixtureWalker(g, opp.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, a) != 4 || loyaltyCount(g, b) != 3 {
		t.Errorf("loyalty after gaining life: %d and %d, want 4 and 3", loyaltyCount(g, a), loyaltyCount(g, b))
	}
	if loyaltyCount(g, theirs) != 2 {
		t.Error("an opponent's planeswalker got a counter")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, a) != 4 {
		t.Error("an opponent gaining life triggered Way of the Mentor")
	}
}

func TestWayOfTheNecromancerGrowsWalkersWhenYourCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Way of the Necromancer", fraEnchantmentLine, fraNecromancerOracle, false)
	jace := seedJaceToken(t, g, me.ID, 3)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, jace) != 3 {
		t.Error("an opponent's creature dying triggered it")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, jace) != 4 {
		t.Errorf("loyalty %d, want 4 after my creature died", loyaltyCount(g, jace))
	}
}

func TestWayOfTheMindSculptorDrawsOnlyForMinusTwoOrLower(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Mind Sculptor", fraEnchantmentLine, fraMindSculptorOracle, false)
	plus := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 5)
	minus := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 5)

	before := handSize(me)
	b16Activate(t, g, me.ID, plus, 0, game.ActivateAbilityParams{}) // +2
	passPriorityAroundTable(t, g)
	if handSize(me) != before {
		t.Fatalf("a +2 drew a card (%d -> %d)", before, handSize(me))
	}
	b16Activate(t, g, me.ID, minus, 1, game.ActivateAbilityParams{}) // −3
	passPriorityAroundTable(t, g)
	if handSize(me) != before+1 {
		t.Fatalf("a −3 drew %d cards, want 1", handSize(me)-before)
	}
}

func TestWayOfTheMindSculptorMinusOneDoesNotDraw(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Mind Sculptor", fraEnchantmentLine, fraMindSculptorOracle, false)
	jace := seedJaceToken(t, g, me.ID, 4)
	before := handSize(me)
	b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{}) // −1 surveil
	passPriorityAroundTable(t, g)
	if handSize(me) != before {
		t.Errorf("a −1 drew a card (%d -> %d)", before, handSize(me))
	}
}

func TestWayOfTheParadoxGainsLifeAndAllowsAnExtraLand(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Way of the Paradox", fraEnchantmentLine, fraParadoxOracle, false)
	jace := seedJaceToken(t, g, me.ID, 4)
	life := me.Life
	b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	playLandFromHand(t, g, "Forest", "")
	playLandFromHand(t, g, "Plains", "")
}

func TestAvatarOfBurgeoningEchoesLandfallEmpowersAndGrantsMinusTen(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Avatar of Burgeoning Echoes", "Creature — Avatar", fraAvatarOracle, false)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	jace := onlyJaceToken(t, g, me.ID)
	if loyaltyCount(g, jace) != 2 {
		t.Fatalf("Jace loyalty %d after landfall, want 2", loyaltyCount(g, jace))
	}

	big := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 10)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	lands := countControlled(g, me.ID, func(c game.Card) bool { return c.IsLand() })
	fraActivateGranted(t, g, me.ID, big, "−10", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if got := plusOneCounters(g, bear); got != lands || lands == 0 {
		t.Errorf("+1/+1 counters = %d, want %d (one per land)", got, lands)
	}
}

func TestSanctumLurkerEmpowersKeepsWalkersAtZeroAndGrantsPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castCatalogSpell(t, g, "Sanctum Lurker", "Creature — Horror", fraLurkerOracle, nil)
	passPriorityAroundTable(t, g)
	jace := onlyJaceToken(t, g, me.ID)
	if loyaltyCount(g, jace) != 1 {
		t.Fatalf("Jace loyalty %d, want 1", loyaltyCount(g, jace))
	}
	// Paid down to 0: survives.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(jace, game.CounterLoyalty, -1) })
	runStateChecks(g)
	if !onBattlefieldNow(g, jace) || loyaltyCount(g, jace) != 0 {
		t.Fatalf("the 0-loyalty Jace should survive (on battlefield %v, loyalty %d)", onBattlefieldNow(g, jace), loyaltyCount(g, jace))
	}
	// The granted +2 pings each opponent and gains 1.
	other := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 1)
	life, oppLife := me.Life, opp.Life
	fraActivateGranted(t, g, me.ID, other, "+2", game.ActivateAbilityParams{})
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("life %d / opp %d, want %d / %d", me.Life, opp.Life, life+1, oppLife-1)
	}
	if loyaltyCount(g, other) != 3 {
		t.Errorf("granted +2 left loyalty %d, want 3", loyaltyCount(g, other))
	}
}

func TestInspiredTethermageGrowsOnLoyaltyCountersAndEmpowers(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := pushCatalogPermanent(g, me.ID, "Inspired Tethermage", "Creature — Elf Warrior", fraTethermageOracle, false)
	b06AddMana(me, "C", "C", "C", "C", "C", "C")
	b16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	jace := onlyJaceToken(t, g, me.ID)
	if loyaltyCount(g, jace) != 2 {
		t.Fatalf("Jace loyalty %d, want 2", loyaltyCount(g, jace))
	}
	if got := plusOneCounters(g, mage); got != 1 {
		t.Errorf("the Tethermage has %d +1/+1 counters, want 1 (one placement)", got)
	}
	w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 0)
	passPriorityAroundTable(t, g)
	base := plusOneCounters(g, mage)
	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{}) // +2
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, mage); got != base+1 {
		t.Errorf("after a +2: %d counters, want %d (one placement)", got, base+1)
	}
}

func TestInspiredTethermageIgnoresLoyaltyRemoval(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// The walker arrives first, so its own arrival is not a placement the
	// Tethermage watches.
	w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 5)
	mage := pushCatalogPermanent(g, me.ID, "Inspired Tethermage", "Creature — Elf Warrior", fraTethermageOracle, false)
	b16Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{}) // −3
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, mage); got != 0 {
		t.Errorf("a −3 gave the Tethermage %d counters, want 0", got)
	}
}

func TestJacesMachinationsOpensInstantSpeedAndEmpowersEight(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Jace's Machinations", "Instant", fraMachinationsOracle, nil)
	passPriorityAroundTable(t, g)
	jace := onlyJaceToken(t, g, me.ID)
	if loyaltyCount(g, jace) != 8 {
		t.Fatalf("Jace loyalty %d, want 8", loyaltyCount(g, jace))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	b16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{}) // −3 outside the main phase
	if loyaltyCount(g, jace) != 5 {
		t.Errorf("the −3 at instant speed left loyalty %d, want 5", loyaltyCount(g, jace))
	}
}

func TestOverwriteTheMultiverseExilesAllAndEmpowersByTheCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := b16Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	castCatalogSpell(t, g, "Overwrite the Multiverse", "Sorcery", fraOverwriteOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if onBattlefieldNow(g, id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if loyaltyCount(g, onlyJaceToken(t, g, me.ID)) != 3 {
		t.Errorf("Jace loyalty %d, want 3", loyaltyCount(g, onlyJaceToken(t, g, me.ID)))
	}
}

func TestOverwriteTheMultiverseWithNoCreaturesMakesNoUsableJace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Overwrite the Multiverse", "Sorcery", fraOverwriteOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range jaceTokensOf(g, me.ID) {
		if loyaltyCount(g, id) > 0 {
			t.Error("a Jace with loyalty appeared from exiling nothing")
		}
	}
}

func TestViolentEchoesEmpowersByTheExcessDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, bear) {
		t.Error("the bear survived 6 damage")
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 4 {
		t.Errorf("Jace loyalty %d, want 4 (6 damage on a 2/2)", got)
	}
}

func TestViolentEchoesWithNoExcessMakesNoJace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := b16Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	castCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, []game.TargetRef{{Kind: game.TargetCard, ID: titan}})
	passPriorityAroundTable(t, g)
	if jaces := jaceTokensOf(g, me.ID); len(jaces) != 0 {
		t.Errorf("no excess damage, yet %d Jace tokens exist", len(jaces))
	}
}

func TestViolentEchoesCountsLoyaltyAsLethalOnAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := fixtureWalker(g, opp.ID, pw2797WalkerOracle, "Planeswalker — Fix", 4)
	castCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, []game.TargetRef{{Kind: game.TargetCard, ID: w}})
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 2 {
		t.Errorf("Jace loyalty %d, want 2 (6 damage on 4 loyalty)", got)
	}
}

func TestHexhavenBattalionMakesThreeCadetsAndEmpowers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Hexhaven Battalion", "Sorcery", fraHexhavenOracle, nil)
	passPriorityAroundTable(t, g)
	if got := fraCountNamed(g, "Cadet"); got != 3 {
		t.Errorf("Cadets = %d, want 3", got)
	}
	if loyaltyCount(g, onlyJaceToken(t, g, me.ID)) != 2 {
		t.Error("Jace is not at 2 loyalty")
	}
}

func TestHexhavenBattalionBasicLandcycles(t *testing.T) {
	spec, ok := Lookup(fraHexhavenOracle)
	if !ok {
		t.Fatal("Hexhaven Battalion is not registered")
	}
	found := false
	for _, ab := range spec.Activated {
		if ab.Label != "" && ab.Cost.Mana == "{2}" {
			found = true
		}
	}
	if !found {
		t.Error("no {2} basic landcycling ability")
	}
}

func TestFateholdCharmModes(t *testing.T) {
	t.Run("draw and empower", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := handSize(me)
		castCatalogSpellWithModes(t, g, "Fatehold Charm", "Instant", fraFateholdCharmOracle, []int{0})
		passPriorityAroundTable(t, g)
		if handSize(me) != before+1 {
			t.Errorf("hand %d -> %d, want +1 (the charm left, one drawn: net 0 from the cast, +1 draw)", before, handSize(me))
		}
		if loyaltyCount(g, onlyJaceToken(t, g, me.ID)) != 2 {
			t.Error("Jace is not at 2 loyalty")
		}
	})
	t.Run("pump", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		castCatalogSpellWithModes(t, g, "Fatehold Charm", "Instant", fraFateholdCharmOracle, []int{2})
		passPriorityAroundTable(t, g)
		if effectivePower(t, g, mine) != 3 || effectiveToughness(t, g, mine) != 4 {
			t.Errorf("my bear is %d/%d, want 3/4", effectivePower(t, g, mine), effectiveToughness(t, g, mine))
		}
		if effectivePower(t, g, theirs) != 2 {
			t.Error("an opponent's creature was pumped")
		}
		if jaces := jaceTokensOf(g, me.ID); len(jaces) != 0 {
			t.Error("the pump mode empowered Jace")
		}
	})
	t.Run("bounce a creature", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Fatehold Charm", TypeLine: "Instant",
			OracleID: fraFateholdCharmOracle, Owner: active.ID, Controller: active.ID})
		toMain(t, g)
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{
			Modes:   []int{1},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if onBattlefieldNow(g, theirs) || !opp.Hand.Contains(theirs) {
			t.Error("the creature was not returned to its owner's hand")
		}
	})
}

func TestTheoristsProxyFlashEmpowersAndProtectsTheNextSpell(t *testing.T) {
	spec, _ := Lookup(fraProxyOracle)
	hasFlash := false
	for _, k := range spec.PrintedKeywords {
		hasFlash = hasFlash || k == "flash"
	}
	if !hasFlash {
		t.Error("Theorist's Proxy does not declare flash")
	}
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Theorist's Proxy", "Creature — Illusion", fraProxyOracle, nil)
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, onlyJaceToken(t, g, me.ID)) != 3 {
		t.Fatalf("Jace loyalty %d, want 3", loyaltyCount(g, onlyJaceToken(t, g, me.ID)))
	}
	proxy := battlefieldIDsNamed(g, "Theorist's Proxy")
	if len(proxy) != 1 {
		t.Fatalf("Proxy on battlefield: %d", len(proxy))
	}
	b06AddMana(me, "U")
	b16Activate(t, g, me.ID, proxy[0], 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, proxy[0]) {
		t.Error("the Proxy was not sacrificed")
	}
	if len(cgShields(g, me)) == 0 {
		t.Error("no can't-be-countered promise was recorded for the next spell")
	}
}

func TestTheoristsSanctumEntersUntappedWithAControlledJace(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedJaceToken(t, g, me.ID, 2)
	playLandFromHand(t, g, "Theorist's Sanctum", fraSanctumOracle)
	ids := battlefieldIDsNamed(g, "Theorist's Sanctum")
	if len(ids) != 1 {
		t.Fatalf("Sanctum on battlefield: %d", len(ids))
	}
	if c, _ := battlefieldCard(g, ids[0]); c.Tapped {
		t.Error("the land entered tapped despite a controlled Jace")
	}
}

func TestTheoristsSanctumEntersTappedWithoutAJace(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	playLandFromHand(t, g, "Theorist's Sanctum", fraSanctumOracle)
	ids := battlefieldIDsNamed(g, "Theorist's Sanctum")
	if len(ids) != 1 {
		t.Fatalf("Sanctum on battlefield: %d", len(ids))
	}
	if c, _ := battlefieldCard(g, ids[0]); !c.Tapped {
		t.Error("the land entered untapped with no Jace and nothing to reveal")
	}
}

func TestTheoristsSanctumActivationEmpowersJace(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Theorist's Sanctum", "Land — Island", fraSanctumOracle, false)
	b06AddMana(me, "C", "C", "U")
	abs := game.ActivatedAbilitiesForCard(cardByID(g, land))
	if len(abs) == 0 {
		t.Fatal("no activated ability")
	}
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, onlyJaceToken(t, g, me.ID)) != 2 {
		t.Errorf("Jace loyalty %d, want 2", loyaltyCount(g, onlyJaceToken(t, g, me.ID)))
	}
}
