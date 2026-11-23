package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monChanges counts EventMonarchChanged events naming `to` as the new
// monarch, from event index `from` on.
func monChanges(g *game.Game, from int, to uuid.UUID) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventMonarchChanged && ev.Actor == to {
			n++
		}
	}
	return n
}

// --- the seam: a card makes you the monarch --------------------------

// TestACardMakesYouTheMonarchAndYouDrawAtYourEndStep is the headline:
// the ETB crowns its controller through SetMonarchForEffect (it would
// have deadlocked through SetMonarch), and CR 725.2's end-step draw
// then pays out as it does for a crown set by hand.
func TestACardMakesYouTheMonarchAndYouDrawAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Court of Grace", "Enchantment", effectstest.CourtOfGraceOracle)
	if g.Monarch != uuid.Nil {
		t.Fatalf("the crown moved before the ETB trigger resolved: %v", g.Monarch)
	}
	effectstest.MonSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want the Court's controller %v", g.Monarch, me.ID)
	}

	before := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.MonSettle(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the monarch drew %d at their end step, want 1", got)
	}
}

// TestAnOpponentStealsACardGivenCrownWithCombatDamage: the crown a card
// handed out moves on combat damage exactly like a manual one.
func TestAnOpponentStealsACardGivenCrownWithCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	// Seat 1 is crowned by an effect (the locked-context path).
	g.WithWriteLock(func() {
		if err := (effects.BecomeTheMonarch{Player: them.ID}).Apply(&effects.Context{Game: g}); err != nil {
			t.Errorf("BecomeTheMonarch: %v", err)
		}
	})
	if g.Monarch != them.ID {
		t.Fatalf("setup: monarch = %v, want %v", g.Monarch, them.ID)
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AttackWith(t, g, them.ID, bear)
	effectstest.MonSettle(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the attacker's controller %v", g.Monarch, me.ID)
	}
}

// --- Custodi Lich: "whenever you become the monarch" -----------------

// TestCustodiLichFiresOnceAndNotAgainWhileYouHoldTheCrown: the Lich's
// own ETB crowns you and its edict fires; a second "you become the
// monarch" while you still hold it is no change of holder (CR 725.3),
// so nothing triggers.
func TestCustodiLichFiresOnceAndNotAgainWhileYouHoldTheCrown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	lich := effectstest.CastAndResolveCreature(t, g, "Custodi Lich", "Creature — Zombie Cleric", effectstest.CustodiLichOracle)
	start := len(g.Events)
	effectstest.MonSettle(t, g) // the ETB resolves; the edict triggers and asks for its target
	effectstest.PickPlayer(t, g, me.ID, them.ID)
	effectstest.MonSettle(t, g)
	effectstest.AnswerSacrifice(t, g, them.ID, victim)
	effectstest.MonSettle(t, g)

	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want %v", g.Monarch, me.ID)
	}
	if z := effectstest.E2Zone(g, victim); z != game.ZoneGraveyard {
		t.Fatalf("the edict's creature is in %q, want the graveyard", z)
	}
	if n := monChanges(g, start, me.ID); n != 1 {
		t.Fatalf("%d monarch changes to you, want 1", n)
	}

	// Become the monarch again while wearing the crown.
	again := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.SetMonarchForEffect(me.ID); err != nil {
			t.Errorf("SetMonarchForEffect: %v", err)
		}
	})
	if n := monChanges(g, again, me.ID); n != 0 {
		t.Errorf("re-crowning the monarch emitted %d changes, want 0", n)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil || effectstest.TriggerOnStack(g, lich) != nil || len(g.PendingTriggers) > 0 {
		t.Error("Custodi Lich triggered on a 'become the monarch' that changed nothing")
	}
}

// TestCustodiLichFiresWhenYouTakeTheCrownBack: any route to the crown
// counts — here the sandbox's manual set, which shares the write.
func TestCustodiLichFiresWhenYouTakeTheCrownBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	lich := effectstest.B43Catalog(g, me.ID, "Custodi Lich", "Creature — Zombie Cleric", effectstest.CustodiLichOracle, 4, 2)
	effectstest.MonCrown(t, g, them.ID)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the Lich triggered on an OPPONENT becoming the monarch")
	}
	effectstest.MonCrown(t, g, me.ID)
	effectstest.MonSettle(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || p.Source != lich {
		t.Fatalf("no Custodi Lich target prompt after you became the monarch: %+v", p)
	}
}

// --- the Courts ------------------------------------------------------

// TestCourtOfGraceUpkeepPaysOffOnlyWhileYouAreTheMonarch: an Angel for
// the monarch, a Spirit otherwise.
func TestCourtOfGraceUpkeepPaysOffOnlyWhileYouAreTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
		want    string
	}{{"monarch", 1, "Angel"}, {"not the monarch", 2, "Spirit"}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			effectstest.PushPermanentForTest(g, owner.ID, "Court of Grace", effectstest.CourtOfGraceOracle, "Enchantment")
			effectstest.MonCrown(t, g, g.Seats[tc.monarch].ID)
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.MonSettle(t, g)
			if n := effectstest.MonTokens(g, owner.ID, tc.want); n != 1 {
				t.Errorf("%d %s tokens, want 1", n, tc.want)
			}
			other := "Spirit"
			if tc.want == "Spirit" {
				other = "Angel"
			}
			if n := effectstest.MonTokens(g, owner.ID, other); n != 0 {
				t.Errorf("%d %s tokens, want 0", n, other)
			}
		})
	}
}

// TestCourtOfIreDealsSevenOnlyAsMonarch.
func TestCourtOfIreDealsSevenOnlyAsMonarch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
		want    int
	}{{"monarch", 1, 7}, {"not the monarch", 3, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner, target := g.Seats[1], g.Seats[2]
			effectstest.PushPermanentForTest(g, owner.ID, "Court of Ire", effectstest.CourtOfIreOracle, "Enchantment")
			effectstest.MonCrown(t, g, g.Seats[tc.monarch].ID)
			before := target.Life
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.MonSettle(t, g)
			effectstest.PickPlayer(t, g, owner.ID, target.ID)
			effectstest.MonSettle(t, g)
			if got := before - target.Life; got != tc.want {
				t.Errorf("Court of Ire dealt %d, want %d", got, tc.want)
			}
		})
	}
}

// TestCourtOfCunningMillsTenForTheMonarch: any number of target
// players, two each or ten each.
func TestCourtOfCunningMillsTenForTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
		want    int
	}{{"monarch", 1, 10}, {"not the monarch", 0, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner, a, b := g.Seats[1], g.Seats[2], g.Seats[3]
			effectstest.PushPermanentForTest(g, owner.ID, "Court of Cunning", effectstest.CourtOfCunningOracle, "Enchantment")
			effectstest.MonCrown(t, g, g.Seats[tc.monarch].ID)
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.MonSettle(t, g)
			aBefore, bBefore, ownBefore := a.Library.Size(), b.Library.Size(), owner.Library.Size()
			p := effectstest.LatestPickTarget(g, owner.ID)
			if p == nil {
				t.Fatal("no target prompt")
			}
			if err := g.ResolvePickTargets(p.ID, owner.ID, []game.TargetRef{
				{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID},
			}); err != nil {
				t.Fatalf("ResolvePickTargets: %v", err)
			}
			effectstest.MonSettle(t, g)
			if got := aBefore - a.Library.Size(); got != tc.want {
				t.Errorf("first target milled %d, want %d", got, tc.want)
			}
			if got := bBefore - b.Library.Size(); got != tc.want {
				t.Errorf("second target milled %d, want %d", got, tc.want)
			}
			if got := ownBefore - owner.Library.Size(); got != 0 {
				t.Errorf("an untargeted player milled %d", got)
			}
		})
	}
}

// TestCourtOfBountyOffersACreatureOnlyToTheMonarch.
func TestCourtOfBountyOffersACreatureOnlyToTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name         string
		monarch      int
		creatureOffe bool
	}{{"monarch", 1, true}, {"not the monarch", 2, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			effectstest.PushPermanentForTest(g, owner.ID, "Court of Bounty", effectstest.CourtOfBountyOracle, "Enchantment")
			land := effectstest.HandCardForTest(owner, "Forest", "Basic Land — Forest", "")
			beast := effectstest.HandCardForTest(owner, "Beast", "Creature — Beast", "")
			effectstest.MonCrown(t, g, g.Seats[tc.monarch].ID)
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.MonSettle(t, g)
			pick := effectstest.ChooseCardsChoiceFor(g, owner.ID)
			if pick == nil {
				t.Fatal("no put-from-hand prompt")
			}
			if !effectstest.HasID(pick.ChooseCards, land) {
				t.Error("the land is not offered")
			}
			if got := effectstest.HasID(pick.ChooseCards, beast); got != tc.creatureOffe {
				t.Errorf("creature offered = %v, want %v", got, tc.creatureOffe)
			}
			effectstest.AnswerChooseCards(t, g, owner.ID, land)
			effectstest.MonSettle(t, g)
			if z := effectstest.E2Zone(g, land); z != game.ZoneBattlefield {
				t.Errorf("the land is in %q, want the battlefield", z)
			}
		})
	}
}

// TestCourtOfAmbitionAsksEachOpponentInTurnOrder: six or two cards for
// the monarch; a player who can't discard two just loses the life.
func TestCourtOfAmbitionAsksEachOpponentInTurnOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, discarder, payer, broke := g.Seats[1], g.Seats[2], g.Seats[3], g.Seats[0]
	effectstest.PushPermanentForTest(g, owner.ID, "Court of Ambition", effectstest.CourtOfAmbitionOracle, "Enchantment")
	effectstest.MonCrown(t, g, owner.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	// Seat 0 holds one card — it cannot pay "discard two cards".
	for broke.Hand.Size() > 1 {
		broke.Hand.Cards = broke.Hand.Cards[:len(broke.Hand.Cards)-1]
	}
	lives := map[uuid.UUID]int{discarder.ID: discarder.Life, payer.ID: payer.Life, broke.ID: broke.Life}
	discarderHand := discarder.Hand.Size()
	effectstest.MonSettle(t, g)

	// APNAP from the active seat (1): seat 2 is asked first, and only
	// seat 2 — the next question waits on this answer.
	if effectstest.LatestOptionPickFor(g, payer.ID) != nil {
		t.Fatal("the second opponent was asked before the first answered")
	}
	effectstest.AnswerOptionPick(t, g, discarder.ID, 1)
	c := effectstest.DiscardChoiceFor(g, discarder.ID)
	if c == nil {
		t.Fatal("no discard prompt after choosing to discard")
	}
	effectstest.AnswerDiscard(t, g, discarder.ID, discarder.Hand.Cards[0].InstanceID, discarder.Hand.Cards[1].InstanceID)
	effectstest.AnswerOptionPick(t, g, payer.ID, 0)
	effectstest.MonSettle(t, g)

	if got := discarderHand - discarder.Hand.Size(); got != 2 {
		t.Errorf("the discarding opponent pitched %d, want 2", got)
	}
	if discarder.Life != lives[discarder.ID] {
		t.Errorf("the discarding opponent lost %d life, want 0", lives[discarder.ID]-discarder.Life)
	}
	if got := lives[payer.ID] - payer.Life; got != 6 {
		t.Errorf("the paying opponent lost %d, want 6", got)
	}
	if got := lives[broke.ID] - broke.Life; got != 6 {
		t.Errorf("the one-card opponent lost %d, want 6 without being asked", got)
	}
	if owner.Life != 40 {
		t.Errorf("the Court's controller lost life: %d", owner.Life)
	}
}

// TestCourtOfAmbitionSmallHalfWithoutTheCrown: three life or one card.
func TestCourtOfAmbitionSmallHalfWithoutTheCrown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, first := g.Seats[1], g.Seats[2]
	effectstest.PushPermanentForTest(g, owner.ID, "Court of Ambition", effectstest.CourtOfAmbitionOracle, "Enchantment")
	effectstest.MonCrown(t, g, g.Seats[3].ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	before := first.Life
	effectstest.MonSettle(t, g)
	c := effectstest.LatestOptionPickFor(g, first.ID)
	if c == nil {
		t.Fatal("no question for the first opponent")
	}
	if got := c.PickOptions[0].LifeCost; got != 3 {
		t.Errorf("the life branch costs %d, want 3", got)
	}
	effectstest.AnswerOptionPick(t, g, first.ID, 0)
	if got := before - first.Life; got != 3 {
		t.Errorf("lost %d, want 3", got)
	}
}

// --- Queen Marchesa: the intervening "if" ----------------------------

func TestQueenMarchesaMakesAnAssassinOnlyWhileAnOpponentIsTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int // -1: nobody
		want    int
	}{{"an opponent", 2, 1}, {"you", 1, 0}, {"nobody", -1, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			queen := effectstest.B43Catalog(g, owner.ID, "Queen Marchesa", "Legendary Creature — Human Assassin", effectstest.QueenMarchesaOracle, 3, 3)
			if tc.monarch >= 0 {
				effectstest.MonCrown(t, g, g.Seats[tc.monarch].ID)
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if tc.want == 0 && effectstest.TriggerOnStack(g, queen) != nil {
				t.Error("the intervening if did not stop the trigger")
			}
			effectstest.MonSettle(t, g)
			if n := effectstest.MonTokens(g, owner.ID, "Assassin"); n != tc.want {
				t.Errorf("%d Assassins, want %d", n, tc.want)
			}
		})
	}
}

// TestQueenMarchesaRechecksOnResolution: taking the crown back with the
// trigger on the stack stops the Assassin (CR 603.4).
func TestQueenMarchesaRechecksOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	queen := effectstest.B43Catalog(g, owner.ID, "Queen Marchesa", "Legendary Creature — Human Assassin", effectstest.QueenMarchesaOracle, 3, 3)
	effectstest.MonCrown(t, g, g.Seats[2].ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, queen) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("setup: no trigger")
	}
	effectstest.MonCrown(t, g, owner.ID)
	effectstest.MonSettle(t, g)
	if n := effectstest.MonTokens(g, owner.ID, "Assassin"); n != 0 {
		t.Errorf("%d Assassins after you took the crown back in response, want 0", n)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, queen, "deathtouch") || !effectstest.EffectiveAbilitiesContain(t, g, queen, "haste") {
		t.Error("Queen Marchesa lacks deathtouch or haste")
	}
}

// --- Knights of the Black Rose ---------------------------------------

func TestKnightsOfTheBlackRoseDrainsTheNewMonarchWhenYouHeldItAtTurnStart(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	effectstest.B43Catalog(g, owner.ID, "Knights of the Black Rose", "Creature — Human Knight", effectstest.KnightsOfTheBlackRoseOracle, 4, 4)
	effectstest.MonCrown(t, g, owner.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1) // seat 1's turn begins with seat 1 wearing the crown
	effectstest.MonSettle(t, g)
	ownerLife, thiefLife := owner.Life, thief.Life

	effectstest.MonCrown(t, g, thief.ID)
	effectstest.MonSettle(t, g)
	if got := thiefLife - thief.Life; got != 2 {
		t.Errorf("the new monarch lost %d, want 2", got)
	}
	if got := owner.Life - ownerLife; got != 2 {
		t.Errorf("you gained %d, want 2", got)
	}
}

func TestKnightsOfTheBlackRoseIgnoresATurnYouDidNotStartAsMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	effectstest.B43Catalog(g, owner.ID, "Knights of the Black Rose", "Creature — Human Knight", effectstest.KnightsOfTheBlackRoseOracle, 4, 4)
	// Seat 0's turn 1 began with nobody wearing the crown.
	effectstest.MonCrown(t, g, owner.ID)
	thiefLife := thief.Life
	effectstest.MonCrown(t, g, thief.ID)
	effectstest.MonSettle(t, g)
	if thief.Life != thiefLife {
		t.Errorf("drained %d on a turn you did not start as the monarch", thiefLife-thief.Life)
	}
}

// --- Entourage of Trest ----------------------------------------------

func TestEntourageOfTrestBlocksAnExtraCreatureOnlyWhileYouAreTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	elf := effectstest.B43Catalog(g, me.ID, "Entourage of Trest", "Creature — Elf Soldier", effectstest.EntourageOfTrestOracle, 4, 4)
	capacity := func() int {
		c := effectstest.E2Card(t, g, elf)
		return game.BlockCapacity(&c)
	}
	if got := capacity(); got != 1 {
		t.Fatalf("block capacity without the crown = %d, want 1", got)
	}
	effectstest.MonCrown(t, g, me.ID)
	if got := capacity(); got != 2 {
		t.Errorf("block capacity as the monarch = %d, want 2", got)
	}
	effectstest.MonCrown(t, g, them.ID)
	if got := capacity(); got != 1 {
		t.Errorf("block capacity after losing the crown = %d, want 1 — the static went stale", got)
	}
}

// --- Regal Behemoth --------------------------------------------------

func TestRegalBehemothAddsAManaOnlyWhileYouAreTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B31Push(g, me.ID, "Regal Behemoth", "Creature — Dinosaur", effectstest.RegalBehemothOracle, "", 5, 5)
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	other := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)

	effectstest.MonCrown(t, g, them.ID)
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 1 {
		t.Fatalf("pool = %v without the crown, want one green", got)
	}
	// The additional mana is a colour pick, so "nothing extra" also
	// means no pick is waiting.
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		t.Fatalf("a colour pick for an additional mana opened without the crown: %+v", pick)
	}

	effectstest.MonCrown(t, g, me.ID)
	effectstest.TapForMana(t, g, me.ID, other)
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick for the additional mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "U" {
		t.Errorf("pool = %v as the monarch, want G G U", got)
	}
}

// --- Throne of the High City -----------------------------------------

func TestThroneOfTheHighCityCrownsYouAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.PushStampedCatalogCard(g, me.ID, "Throne of the High City", "Land", effectstest.ThroneOfTheHighCityOracle)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	for i := 0; i < 4; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if z := effectstest.E2Zone(g, throne); z != game.ZoneGraveyard {
		t.Errorf("the Throne is in %q after paying its cost, want the graveyard", z)
	}
	if g.Monarch == me.ID {
		t.Fatal("the crown moved before the ability resolved")
	}
	effectstest.MonSettle(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want %v", g.Monarch, me.ID)
	}
}

func TestThroneOfTheHighCityCannotBeActivatedWithoutTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.PushStampedCatalogCard(g, me.ID, "Throne of the High City", "Land", effectstest.ThroneOfTheHighCityOracle)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {4}, {T}, Sacrifice with an empty pool")
	}
	if g.Monarch != uuid.Nil || effectstest.E2Zone(g, throne) != game.ZoneBattlefield {
		t.Error("a refused activation moved the crown or the land")
	}
}

func TestPalaceJailerReturnsTheCreatureWhenAnOpponentBecomesTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, owner, thief := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := effectstest.PushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailer := effectstest.JailerExile(t, g, victim)

	// Neither the Jailer leaving, nor the turn ending, nor the crown
	// being cleared, nor YOU becoming the monarch releases it: the
	// duration is keyed to an OPPONENT taking the crown.
	effectstest.E2Destroy(t, g, jailer)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.MonSettle(t, g)
	effectstest.MonCrown(t, g, uuid.Nil)
	effectstest.MonCrown(t, g, me.ID)
	effectstest.MonSettle(t, g)
	if z := effectstest.E2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature left exile early (in %q)", z)
	}

	effectstest.MonCrown(t, g, thief.ID)
	effectstest.MonSettle(t, g)
	back, ok := effectstest.E2BattlefieldNamed(g, "Their Bear")
	if !ok {
		t.Fatal("an opponent became the monarch and the creature did not return")
	}
	if back.Controller != owner.ID {
		t.Errorf("returned under %v, want its owner %v", back.Controller, owner.ID)
	}
	// It fires once: the crown moving again returns nothing twice.
	effectstest.MonCrown(t, g, owner.ID)
	effectstest.MonSettle(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d copies of the creature on the battlefield, want 1", n)
	}
}

func TestPalaceJailerTargetsOnlyAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Palace Jailer", "Creature — Human Soldier", effectstest.PalaceJailerOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("your own creature is offered")
	}
	if !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("the opponent's creature is not offered")
	}
}

// TestPalaceJailerExileSurvivesASnapshotRoundTrip: the release is an
// "until" record (CR 610.3, #1729) — data — so a table with a creature
// jailed is a restore point, and the restored game still releases it.
// The controller leaving the game is in until_return_test.go.
func TestPalaceJailerExileSurvivesASnapshotRoundTrip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	victim := effectstest.PushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	effectstest.JailerExile(t, g, victim)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if restored.Monarch != g.Seats[0].ID {
		t.Fatalf("restored monarch = %v, want %v", restored.Monarch, g.Seats[0].ID)
	}
	effectstest.MonCrown(t, restored, thief.ID)
	effectstest.MonSettle(t, restored)
	if _, ok := effectstest.E2BattlefieldNamed(restored, "Their Bear"); !ok {
		t.Error("the restored game did not release the creature")
	}
}

// TestPalaceJailerExileSurvivesUndo: Clone / RestoreFrom keeps the
// until record and the crown.
func TestPalaceJailerExileSurvivesUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	victim := effectstest.PushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	effectstest.JailerExile(t, g, victim)

	before := g.Clone()
	effectstest.MonCrown(t, g, thief.ID)
	effectstest.MonSettle(t, g)
	g.RestoreFrom(before)
	if g.Monarch != g.Seats[0].ID {
		t.Fatalf("after undo monarch = %v, want %v", g.Monarch, g.Seats[0].ID)
	}
	if z := effectstest.E2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("after undo the creature is in %q, want exile", z)
	}
	effectstest.MonCrown(t, g, thief.ID)
	effectstest.MonSettle(t, g)
	if _, ok := effectstest.E2BattlefieldNamed(g, "Their Bear"); !ok {
		t.Error("after undo the crown moved and the creature did not return")
	}
}
