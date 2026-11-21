package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	idx, ref := effectstest.GrantedLoyaltyRow(t, g, walker, prefix)
	params.Ref = ref
	if err := g.ActivateCatalogAbility(who, walker, idx, params); err != nil {
		t.Fatalf("activate %q: %v", prefix, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func fraCountNamed(g *game.Game, name string) int { return effectstest.B16CountNamed(g, name) }

// castWay casts a Way, which enters and empowers Jace n.
func fraCastWay(t *testing.T, g *game.Game, name, oracle string) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, fraEnchantmentLine, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			fraCastWay(t, g, tc.name, tc.oracle)
			if got := effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)); got != tc.n {
				t.Fatalf("Jace loyalty %d, want %d", got, tc.n)
			}
		})
	}
}

func TestWayOfTheDeathbringerSacrificesForABeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Deathbringer", fraEnchantmentLine, fraDeathbringerOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)

	fraActivateGranted(t, g, me.ID, jace, "−2", game.ActivateAbilityParams{})
	if got := effectstest.LoyaltyCount(g, jace); got != 2 {
		t.Fatalf("loyalty after −2: %d, want 2", got)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	beasts := effectstest.BattlefieldIDsNamed(g, "Beast")
	if len(beasts) != 1 {
		t.Fatalf("Beasts = %d, want 1", len(beasts))
	}
	if effectstest.EffectivePower(t, g, beasts[0]) != 4 || !effectstest.EffectiveAbilitiesContain(t, g, beasts[0], "trample") {
		t.Error("the Beast is not a 4/4 with trample")
	}
}

func TestWayOfTheDeathbringerDeclinedMakesNoBeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Deathbringer", fraEnchantmentLine, fraDeathbringerOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)

	fraActivateGranted(t, g, me.ID, jace, "−2", game.ActivateAbilityParams{})
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldNow(g, bear) || fraCountNamed(g, "Beast") != 0 {
		t.Error("declining the sacrifice still made a Beast or lost the creature")
	}
}

func TestWayOfTheWildspeakerMakesABeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Wildspeaker", fraEnchantmentLine, fraWildspeakerOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)
	short := effectstest.SeedJaceToken(t, g, me.ID, 3)

	idx, ref := effectstest.GrantedLoyaltyRow(t, g, short, "−4")
	if err := g.ActivateCatalogAbility(me.ID, short, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−4 with 3 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	fraActivateGranted(t, g, me.ID, jace, "−4", game.ActivateAbilityParams{})
	beasts := effectstest.BattlefieldIDsNamed(g, "Beast")
	if len(beasts) != 1 || effectstest.EffectivePower(t, g, beasts[0]) != 4 || !effectstest.EffectiveAbilitiesContain(t, g, beasts[0], "trample") {
		t.Fatalf("want one 4/4 trampling Beast, got %v", beasts)
	}
}

func TestWayOfTheHealerMakesACadetAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Healer", fraEnchantmentLine, fraHealerOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)

	fraActivateGranted(t, g, me.ID, jace, "−2", game.ActivateAbilityParams{})
	cadets := effectstest.BattlefieldIDsNamed(g, "Cadet")
	if len(cadets) != 1 {
		t.Fatalf("Cadets = %d, want 1", len(cadets))
	}
	c := effectstest.CardByID(g, cadets[0])
	if c.TypeLine != "Token Creature — Wizard Soldier" || effectstest.EffectivePower(t, g, cadets[0]) != 2 || len(c.Colors) != 0 {
		t.Errorf("Cadet = %q %d power colours %v", c.TypeLine, effectstest.EffectivePower(t, g, cadets[0]), c.Colors)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("no surveil prompt was queued")
	}
}

func TestWayOfThePyromancerAddsRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Pyromancer", fraEnchantmentLine, fraPyromancerOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 1)

	fraActivateGranted(t, g, me.ID, jace, "+1", game.ActivateAbilityParams{})
	if got := effectstest.LoyaltyCount(g, jace); got != 2 {
		t.Errorf("loyalty after +1: %d, want 2", got)
	}
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool = %d, want one red", got)
	}
}

func TestWayOfTheWarlordHitsCreatureAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Warlord", fraEnchantmentLine, fraWarlordOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 5)
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life

	fraActivateGranted(t, g, me.ID, jace, "−4", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if effectstest.OnBattlefieldNow(g, bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

func TestWayOfTheWarlordNoCreatureStillHitsThePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Warlord", fraEnchantmentLine, fraWarlordOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 5)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Cryomancer", fraEnchantmentLine, fraCryoOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)
	fraActivateGranted(t, g, me.ID, jace, "−3", game.ActivateAbilityParams{})

	handBefore := effectstest.HandSize(me)
	// An instant that draws a card: the copy draws a second.
	effectstest.CastCatalogSpell(t, g, "Fixture Draw", "Instant", effectstest.FraDrawOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// A copy trigger may offer new targets; none here.
	effectstest.PassPriorityAroundTable(t, g)
	// one card left the hand (the cast), two drawn
	if got := effectstest.HandSize(me); got != handBefore+1 && got != handBefore+2 {
		t.Fatalf("hand %d -> %d after a copied draw spell", handBefore, got)
	}
}

func TestWayOfTheMentorGrowsWalkersOnLifegain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Mentor", fraEnchantmentLine, fraMentorOracle, false)
	a := effectstest.SeedJaceToken(t, g, me.ID, 3)
	b := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)
	theirs := effectstest.FixtureWalker(g, opp.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, a) != 4 || effectstest.LoyaltyCount(g, b) != 3 {
		t.Errorf("loyalty after gaining life: %d and %d, want 4 and 3", effectstest.LoyaltyCount(g, a), effectstest.LoyaltyCount(g, b))
	}
	if effectstest.LoyaltyCount(g, theirs) != 2 {
		t.Error("an opponent's planeswalker got a counter")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, a) != 4 {
		t.Error("an opponent gaining life triggered Way of the Mentor")
	}
}

func TestWayOfTheNecromancerGrowsWalkersWhenYourCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Necromancer", fraEnchantmentLine, fraNecromancerOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 3)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, jace) != 3 {
		t.Error("an opponent's creature dying triggered it")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, jace) != 4 {
		t.Errorf("loyalty %d, want 4 after my creature died", effectstest.LoyaltyCount(g, jace))
	}
}

func TestWayOfTheMindSculptorDrawsOnlyForMinusTwoOrLower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Mind Sculptor", fraEnchantmentLine, fraMindSculptorOracle, false)
	plus := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 5)
	minus := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 5)

	before := effectstest.HandSize(me)
	effectstest.B16Activate(t, g, me.ID, plus, 0, game.ActivateAbilityParams{}) // +2
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != before {
		t.Fatalf("a +2 drew a card (%d -> %d)", before, effectstest.HandSize(me))
	}
	effectstest.B16Activate(t, g, me.ID, minus, 1, game.ActivateAbilityParams{}) // −3
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != before+1 {
		t.Fatalf("a −3 drew %d cards, want 1", effectstest.HandSize(me)-before)
	}
}

func TestWayOfTheMindSculptorMinusOneDoesNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Mind Sculptor", fraEnchantmentLine, fraMindSculptorOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)
	before := effectstest.HandSize(me)
	effectstest.B16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{}) // −1 surveil
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != before {
		t.Errorf("a −1 drew a card (%d -> %d)", before, effectstest.HandSize(me))
	}
}

func TestWayOfTheParadoxGainsLifeAndAllowsAnExtraLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Way of the Paradox", fraEnchantmentLine, fraParadoxOracle, false)
	jace := effectstest.SeedJaceToken(t, g, me.ID, 4)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PlayLandFromHand(t, g, "Plains", "")
}

func TestAvatarOfBurgeoningEchoesLandfallEmpowersAndGrantsMinusTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Avatar of Burgeoning Echoes", "Creature — Avatar", fraAvatarOracle, false)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	jace := effectstest.OnlyJaceToken(t, g, me.ID)
	if effectstest.LoyaltyCount(g, jace) != 2 {
		t.Fatalf("Jace loyalty %d after landfall, want 2", effectstest.LoyaltyCount(g, jace))
	}

	big := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 10)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	lands := effects.CountControlled(g, me.ID, func(c game.Card) bool { return c.IsLand() })
	fraActivateGranted(t, g, me.ID, big, "−10", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if got := effectstest.PlusOneCounters(g, bear); got != lands || lands == 0 {
		t.Errorf("+1/+1 counters = %d, want %d (one per land)", got, lands)
	}
}

func TestSanctumLurkerEmpowersKeepsWalkersAtZeroAndGrantsPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastCatalogSpell(t, g, "Sanctum Lurker", "Creature — Horror", fraLurkerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	jace := effectstest.OnlyJaceToken(t, g, me.ID)
	if effectstest.LoyaltyCount(g, jace) != 1 {
		t.Fatalf("Jace loyalty %d, want 1", effectstest.LoyaltyCount(g, jace))
	}
	// Paid down to 0: survives.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(jace, game.CounterLoyalty, -1) })
	effectstest.RunStateChecks(g)
	if !effectstest.OnBattlefieldNow(g, jace) || effectstest.LoyaltyCount(g, jace) != 0 {
		t.Fatalf("the 0-loyalty Jace should survive (on battlefield %v, loyalty %d)", effectstest.OnBattlefieldNow(g, jace), effectstest.LoyaltyCount(g, jace))
	}
	// The granted +2 pings each opponent and gains 1.
	other := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 1)
	life, oppLife := me.Life, opp.Life
	fraActivateGranted(t, g, me.ID, other, "+2", game.ActivateAbilityParams{})
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("life %d / opp %d, want %d / %d", me.Life, opp.Life, life+1, oppLife-1)
	}
	if effectstest.LoyaltyCount(g, other) != 3 {
		t.Errorf("granted +2 left loyalty %d, want 3", effectstest.LoyaltyCount(g, other))
	}
}

func TestInspiredTethermageGrowsOnLoyaltyCountersAndEmpowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Inspired Tethermage", "Creature — Elf Warrior", fraTethermageOracle, false)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	jace := effectstest.OnlyJaceToken(t, g, me.ID)
	if effectstest.LoyaltyCount(g, jace) != 2 {
		t.Fatalf("Jace loyalty %d, want 2", effectstest.LoyaltyCount(g, jace))
	}
	if got := effectstest.PlusOneCounters(g, mage); got != 1 {
		t.Errorf("the Tethermage has %d +1/+1 counters, want 1 (one placement)", got)
	}
	w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 0)
	effectstest.PassPriorityAroundTable(t, g)
	base := effectstest.PlusOneCounters(g, mage)
	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{}) // +2
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, mage); got != base+1 {
		t.Errorf("after a +2: %d counters, want %d (one placement)", got, base+1)
	}
}

func TestInspiredTethermageIgnoresLoyaltyRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// The walker arrives first, so its own arrival is not a placement the
	// Tethermage watches.
	w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 5)
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Inspired Tethermage", "Creature — Elf Warrior", fraTethermageOracle, false)
	effectstest.B16Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{}) // −3
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, mage); got != 0 {
		t.Errorf("a −3 gave the Tethermage %d counters, want 0", got)
	}
}

func TestJacesMachinationsOpensInstantSpeedAndEmpowersEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Jace's Machinations", "Instant", fraMachinationsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	jace := effectstest.OnlyJaceToken(t, g, me.ID)
	if effectstest.LoyaltyCount(g, jace) != 8 {
		t.Fatalf("Jace loyalty %d, want 8", effectstest.LoyaltyCount(g, jace))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.B16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{}) // −3 outside the main phase
	if effectstest.LoyaltyCount(g, jace) != 5 {
		t.Errorf("the −3 at instant speed left loyalty %d, want 5", effectstest.LoyaltyCount(g, jace))
	}
}

func TestOverwriteTheMultiverseExilesAllAndEmpowersByTheCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := effectstest.B16Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Overwrite the Multiverse", "Sorcery", fraOverwriteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if effectstest.OnBattlefieldNow(g, id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)) != 3 {
		t.Errorf("Jace loyalty %d, want 3", effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)))
	}
}

func TestOverwriteTheMultiverseWithNoCreaturesMakesNoUsableJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Overwrite the Multiverse", "Sorcery", fraOverwriteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range effectstest.JaceTokensOf(g, me.ID) {
		if effectstest.LoyaltyCount(g, id) > 0 {
			t.Error("a Jace with loyalty appeared from exiling nothing")
		}
	}
}

func TestViolentEchoesEmpowersByTheExcessDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, bear) {
		t.Error("the bear survived 6 damage")
	}
	if got := effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)); got != 4 {
		t.Errorf("Jace loyalty %d, want 4 (6 damage on a 2/2)", got)
	}
}

func TestViolentEchoesWithNoExcessMakesNoJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.B16Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, []game.TargetRef{{Kind: game.TargetCard, ID: titan}})
	effectstest.PassPriorityAroundTable(t, g)
	if jaces := effectstest.JaceTokensOf(g, me.ID); len(jaces) != 0 {
		t.Errorf("no excess damage, yet %d Jace tokens exist", len(jaces))
	}
}

func TestViolentEchoesCountsLoyaltyAsLethalOnAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.FixtureWalker(g, opp.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 4)
	effectstest.CastCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, []game.TargetRef{{Kind: game.TargetCard, ID: w}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)); got != 2 {
		t.Errorf("Jace loyalty %d, want 2 (6 damage on 4 loyalty)", got)
	}
}

func TestHexhavenBattalionMakesThreeCadetsAndEmpowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Hexhaven Battalion", "Sorcery", fraHexhavenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := fraCountNamed(g, "Cadet"); got != 3 {
		t.Errorf("Cadets = %d, want 3", got)
	}
	if effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)) != 2 {
		t.Error("Jace is not at 2 loyalty")
	}
}

func TestHexhavenBattalionBasicLandcycles(t *testing.T) {
	spec, ok := effects.Lookup(fraHexhavenOracle)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := effectstest.HandSize(me)
		effectstest.CastCatalogSpellWithModes(t, g, "Fatehold Charm", "Instant", fraFateholdCharmOracle, []int{0})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HandSize(me) != before+1 {
			t.Errorf("hand %d -> %d, want +1 (the charm left, one drawn: net 0 from the cast, +1 draw)", before, effectstest.HandSize(me))
		}
		if effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)) != 2 {
			t.Error("Jace is not at 2 loyalty")
		}
	})
	t.Run("pump", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpellWithModes(t, g, "Fatehold Charm", "Instant", fraFateholdCharmOracle, []int{2})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.EffectivePower(t, g, mine) != 3 || effectstest.EffectiveToughness(t, g, mine) != 4 {
			t.Errorf("my bear is %d/%d, want 3/4", effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine))
		}
		if effectstest.EffectivePower(t, g, theirs) != 2 {
			t.Error("an opponent's creature was pumped")
		}
		if jaces := effectstest.JaceTokensOf(g, me.ID); len(jaces) != 0 {
			t.Error("the pump mode empowered Jace")
		}
	})
	t.Run("bounce a creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Fatehold Charm", TypeLine: "Instant",
			OracleID: fraFateholdCharmOracle, Owner: active.ID, Controller: active.ID})
		effectstest.ToMain(t, g)
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{
			Modes:   []int{1},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldNow(g, theirs) || !opp.Hand.Contains(theirs) {
			t.Error("the creature was not returned to its owner's hand")
		}
	})
}

func TestTheoristsProxyFlashEmpowersAndProtectsTheNextSpell(t *testing.T) {
	spec, _ := effects.Lookup(fraProxyOracle)
	hasFlash := false
	for _, k := range spec.PrintedKeywords {
		hasFlash = hasFlash || k == "flash"
	}
	if !hasFlash {
		t.Error("Theorist's Proxy does not declare flash")
	}
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Theorist's Proxy", "Creature — Illusion", fraProxyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)) != 3 {
		t.Fatalf("Jace loyalty %d, want 3", effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)))
	}
	proxy := effectstest.BattlefieldIDsNamed(g, "Theorist's Proxy")
	if len(proxy) != 1 {
		t.Fatalf("Proxy on battlefield: %d", len(proxy))
	}
	effectstest.B06AddMana(me, "U")
	effectstest.B16Activate(t, g, me.ID, proxy[0], 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, proxy[0]) {
		t.Error("the Proxy was not sacrificed")
	}
	if len(effectstest.CgShields(g, me)) == 0 {
		t.Error("no can't-be-countered promise was recorded for the next spell")
	}
}

func TestTheoristsSanctumEntersUntappedWithAControlledJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedJaceToken(t, g, me.ID, 2)
	effectstest.PlayLandFromHand(t, g, "Theorist's Sanctum", fraSanctumOracle)
	ids := effectstest.BattlefieldIDsNamed(g, "Theorist's Sanctum")
	if len(ids) != 1 {
		t.Fatalf("Sanctum on battlefield: %d", len(ids))
	}
	if c, _ := effectstest.BattlefieldCard(g, ids[0]); c.Tapped {
		t.Error("the land entered tapped despite a controlled Jace")
	}
}

func TestTheoristsSanctumEntersTappedWithoutAJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	effectstest.PlayLandFromHand(t, g, "Theorist's Sanctum", fraSanctumOracle)
	ids := effectstest.BattlefieldIDsNamed(g, "Theorist's Sanctum")
	if len(ids) != 1 {
		t.Fatalf("Sanctum on battlefield: %d", len(ids))
	}
	if c, _ := effectstest.BattlefieldCard(g, ids[0]); !c.Tapped {
		t.Error("the land entered untapped with no Jace and nothing to reveal")
	}
}

func TestTheoristsSanctumActivationEmpowersJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Theorist's Sanctum", "Land — Island", fraSanctumOracle, false)
	effectstest.B06AddMana(me, "C", "C", "U")
	abs := game.ActivatedAbilitiesForCard(effectstest.CardByID(g, land))
	if len(abs) == 0 {
		t.Fatal("no activated ability")
	}
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)) != 2 {
		t.Errorf("Jace loyalty %d, want 2", effectstest.LoyaltyCount(g, effectstest.OnlyJaceToken(t, g, me.ID)))
	}
}
