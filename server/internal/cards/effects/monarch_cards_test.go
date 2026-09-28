package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monarch_cards_test.go pins #1722: a card effect can make a player
// the monarch (CR 724), and the twelve catalog cards built on it.

const (
	courtOfGraceOracle          = "f63c2438-27d4-449a-828f-f0a2ea86ff16"
	courtOfAmbitionOracle       = "652e71a9-e46f-41b3-8695-76b3606b1955"
	courtOfCunningOracle        = "2ab88c99-aaa0-4a91-9225-0bbfba04b6bc"
	courtOfBountyOracle         = "e55d9377-f89a-41e5-a094-730d6f24caf0"
	courtOfIreOracle            = "13f58292-9b78-4cd1-a16e-b1779a170d33"
	queenMarchesaOracle         = "d7ac4be1-dcca-49b6-8ddb-d1b0e6cf2dcf"
	palaceJailerOracle          = "180eda7c-fca2-403b-85cd-8ffebaf9f408"
	entourageOfTrestOracle      = "27b0bd3f-2c11-474a-8512-fe035a514905"
	custodiLichOracle           = "0d95ee14-9ff0-4d7f-be56-361a500ce36f"
	knightsOfTheBlackRoseOracle = "9dbfa01b-ce7e-4dd1-9257-0c89b831f477"
	regalBehemothOracle         = "797c2b1c-c373-4735-b397-f561559c7c61"
	throneOfTheHighCityOracle   = "9684447a-5955-4bc7-8ad0-8bb8b316873b"
)

// monSettle resolves everything on and headed for the stack, answering
// any trigger-ordering prompt in the order it offers. It stops at any
// other prompt, which the caller answers.
func monSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) || len(g.PendingChoices) > 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("the stack did not settle")
}

// monCrown sets the monarch by hand (the sandbox action), for setup.
func monCrown(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	if err := g.SetMonarch(player); err != nil {
		t.Fatalf("SetMonarch: %v", err)
	}
}

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

// monTokens counts `controller`'s battlefield tokens named `name`.
func monTokens(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && IsToken(c) {
			n++
		}
	}
	return n
}

// --- the seam: a card makes you the monarch --------------------------

// TestACardMakesYouTheMonarchAndYouDrawAtYourEndStep is the headline:
// the ETB crowns its controller through SetMonarchForEffect (it would
// have deadlocked through SetMonarch), and CR 724.2's end-step draw
// then pays out as it does for a crown set by hand.
func TestACardMakesYouTheMonarchAndYouDrawAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Court of Grace", "Enchantment", courtOfGraceOracle)
	if g.Monarch != uuid.Nil {
		t.Fatalf("the crown moved before the ETB trigger resolved: %v", g.Monarch)
	}
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want the Court's controller %v", g.Monarch, me.ID)
	}

	before := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	monSettle(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the monarch drew %d at their end step, want 1", got)
	}
}

// TestAnOpponentStealsACardGivenCrownWithCombatDamage: the crown a card
// handed out moves on combat damage exactly like a manual one.
func TestAnOpponentStealsACardGivenCrownWithCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	// Seat 1 is crowned by an effect (the locked-context path).
	g.WithWriteLock(func() {
		if err := (BecomeTheMonarch{Player: them.ID}).Apply(&Context{Game: g}); err != nil {
			t.Errorf("BecomeTheMonarch: %v", err)
		}
	})
	if g.Monarch != them.ID {
		t.Fatalf("setup: monarch = %v, want %v", g.Monarch, them.ID)
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	attackWith(t, g, them.ID, bear)
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the attacker's controller %v", g.Monarch, me.ID)
	}
}

// --- Custodi Lich: "whenever you become the monarch" -----------------

// TestCustodiLichFiresOnceAndNotAgainWhileYouHoldTheCrown: the Lich's
// own ETB crowns you and its edict fires; a second "you become the
// monarch" while you still hold it is no change of holder (CR 724.3),
// so nothing triggers.
func TestCustodiLichFiresOnceAndNotAgainWhileYouHoldTheCrown(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	lich := castAndResolveCreature(t, g, "Custodi Lich", "Creature — Zombie Cleric", custodiLichOracle)
	start := len(g.Events)
	monSettle(t, g) // the ETB resolves; the edict triggers and asks for its target
	pickPlayer(t, g, me.ID, them.ID)
	monSettle(t, g)
	answerSacrifice(t, g, them.ID, victim)
	monSettle(t, g)

	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want %v", g.Monarch, me.ID)
	}
	if z := e2Zone(g, victim); z != game.ZoneGraveyard {
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
	if p := latestPickTarget(g, me.ID); p != nil || triggerOnStack(g, lich) != nil || len(g.PendingTriggers) > 0 {
		t.Error("Custodi Lich triggered on a 'become the monarch' that changed nothing")
	}
}

// TestCustodiLichFiresWhenYouTakeTheCrownBack: any route to the crown
// counts — here the sandbox's manual set, which shares the write.
func TestCustodiLichFiresWhenYouTakeTheCrownBack(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	lich := b43Catalog(g, me.ID, "Custodi Lich", "Creature — Zombie Cleric", custodiLichOracle, 4, 2)
	monCrown(t, g, them.ID)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the Lich triggered on an OPPONENT becoming the monarch")
	}
	monCrown(t, g, me.ID)
	monSettle(t, g)
	if p := latestPickTarget(g, me.ID); p == nil || p.Source != lich {
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
			g := newCatalogGame(t)
			owner := g.Seats[1]
			pushPermanentForTest(g, owner.ID, "Court of Grace", courtOfGraceOracle, "Enchantment")
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			if n := monTokens(g, owner.ID, tc.want); n != 1 {
				t.Errorf("%d %s tokens, want 1", n, tc.want)
			}
			other := "Spirit"
			if tc.want == "Spirit" {
				other = "Angel"
			}
			if n := monTokens(g, owner.ID, other); n != 0 {
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
			g := newCatalogGame(t)
			owner, target := g.Seats[1], g.Seats[2]
			pushPermanentForTest(g, owner.ID, "Court of Ire", courtOfIreOracle, "Enchantment")
			monCrown(t, g, g.Seats[tc.monarch].ID)
			before := target.Life
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			pickPlayer(t, g, owner.ID, target.ID)
			monSettle(t, g)
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
			g := newCatalogGame(t)
			owner, a, b := g.Seats[1], g.Seats[2], g.Seats[3]
			pushPermanentForTest(g, owner.ID, "Court of Cunning", courtOfCunningOracle, "Enchantment")
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			aBefore, bBefore, ownBefore := a.Library.Size(), b.Library.Size(), owner.Library.Size()
			p := latestPickTarget(g, owner.ID)
			if p == nil {
				t.Fatal("no target prompt")
			}
			if err := g.ResolvePickTargets(p.ID, owner.ID, []game.TargetRef{
				{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID},
			}); err != nil {
				t.Fatalf("ResolvePickTargets: %v", err)
			}
			monSettle(t, g)
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
			g := newCatalogGame(t)
			owner := g.Seats[1]
			pushPermanentForTest(g, owner.ID, "Court of Bounty", courtOfBountyOracle, "Enchantment")
			land := handCardForTest(owner, "Forest", "Basic Land — Forest", "")
			beast := handCardForTest(owner, "Beast", "Creature — Beast", "")
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			pick := chooseCardsChoiceFor(g, owner.ID)
			if pick == nil {
				t.Fatal("no put-from-hand prompt")
			}
			if !hasID(pick.ChooseCards, land) {
				t.Error("the land is not offered")
			}
			if got := hasID(pick.ChooseCards, beast); got != tc.creatureOffe {
				t.Errorf("creature offered = %v, want %v", got, tc.creatureOffe)
			}
			answerChooseCards(t, g, owner.ID, land)
			monSettle(t, g)
			if z := e2Zone(g, land); z != game.ZoneBattlefield {
				t.Errorf("the land is in %q, want the battlefield", z)
			}
		})
	}
}

// TestCourtOfAmbitionAsksEachOpponentInTurnOrder: six or two cards for
// the monarch; a player who can't discard two just loses the life.
func TestCourtOfAmbitionAsksEachOpponentInTurnOrder(t *testing.T) {
	g := newCatalogGame(t)
	owner, discarder, payer, broke := g.Seats[1], g.Seats[2], g.Seats[3], g.Seats[0]
	pushPermanentForTest(g, owner.ID, "Court of Ambition", courtOfAmbitionOracle, "Enchantment")
	monCrown(t, g, owner.ID)
	advanceToUpkeepOf(t, g, 1)
	// Seat 0 holds one card — it cannot pay "discard two cards".
	for broke.Hand.Size() > 1 {
		broke.Hand.Cards = broke.Hand.Cards[:len(broke.Hand.Cards)-1]
	}
	lives := map[uuid.UUID]int{discarder.ID: discarder.Life, payer.ID: payer.Life, broke.ID: broke.Life}
	discarderHand := discarder.Hand.Size()
	monSettle(t, g)

	// APNAP from the active seat (1): seat 2 is asked first, and only
	// seat 2 — the next question waits on this answer.
	if latestOptionPickFor(g, payer.ID) != nil {
		t.Fatal("the second opponent was asked before the first answered")
	}
	answerOptionPick(t, g, discarder.ID, 1)
	c := discardChoiceFor(g, discarder.ID)
	if c == nil {
		t.Fatal("no discard prompt after choosing to discard")
	}
	answerDiscard(t, g, discarder.ID, discarder.Hand.Cards[0].InstanceID, discarder.Hand.Cards[1].InstanceID)
	answerOptionPick(t, g, payer.ID, 0)
	monSettle(t, g)

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
	g := newCatalogGame(t)
	owner, first := g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, owner.ID, "Court of Ambition", courtOfAmbitionOracle, "Enchantment")
	monCrown(t, g, g.Seats[3].ID)
	advanceToUpkeepOf(t, g, 1)
	before := first.Life
	monSettle(t, g)
	c := latestOptionPickFor(g, first.ID)
	if c == nil {
		t.Fatal("no question for the first opponent")
	}
	if got := c.PickOptions[0].LifeCost; got != 3 {
		t.Errorf("the life branch costs %d, want 3", got)
	}
	answerOptionPick(t, g, first.ID, 0)
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
			g := newCatalogGame(t)
			owner := g.Seats[1]
			queen := b43Catalog(g, owner.ID, "Queen Marchesa", "Legendary Creature — Human Assassin", queenMarchesaOracle, 3, 3)
			if tc.monarch >= 0 {
				monCrown(t, g, g.Seats[tc.monarch].ID)
			}
			advanceToUpkeepOf(t, g, 1)
			if tc.want == 0 && triggerOnStack(g, queen) != nil {
				t.Error("the intervening if did not stop the trigger")
			}
			monSettle(t, g)
			if n := monTokens(g, owner.ID, "Assassin"); n != tc.want {
				t.Errorf("%d Assassins, want %d", n, tc.want)
			}
		})
	}
}

// TestQueenMarchesaRechecksOnResolution: taking the crown back with the
// trigger on the stack stops the Assassin (CR 603.4).
func TestQueenMarchesaRechecksOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	queen := b43Catalog(g, owner.ID, "Queen Marchesa", "Legendary Creature — Human Assassin", queenMarchesaOracle, 3, 3)
	monCrown(t, g, g.Seats[2].ID)
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, queen) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("setup: no trigger")
	}
	monCrown(t, g, owner.ID)
	monSettle(t, g)
	if n := monTokens(g, owner.ID, "Assassin"); n != 0 {
		t.Errorf("%d Assassins after you took the crown back in response, want 0", n)
	}
	if !effectiveAbilitiesContain(t, g, queen, "deathtouch") || !effectiveAbilitiesContain(t, g, queen, "haste") {
		t.Error("Queen Marchesa lacks deathtouch or haste")
	}
}

// --- Knights of the Black Rose ---------------------------------------

func TestKnightsOfTheBlackRoseDrainsTheNewMonarchWhenYouHeldItAtTurnStart(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	b43Catalog(g, owner.ID, "Knights of the Black Rose", "Creature — Human Knight", knightsOfTheBlackRoseOracle, 4, 4)
	monCrown(t, g, owner.ID)
	advanceToUpkeepOf(t, g, 1) // seat 1's turn begins with seat 1 wearing the crown
	monSettle(t, g)
	ownerLife, thiefLife := owner.Life, thief.Life

	monCrown(t, g, thief.ID)
	monSettle(t, g)
	if got := thiefLife - thief.Life; got != 2 {
		t.Errorf("the new monarch lost %d, want 2", got)
	}
	if got := owner.Life - ownerLife; got != 2 {
		t.Errorf("you gained %d, want 2", got)
	}
}

func TestKnightsOfTheBlackRoseIgnoresATurnYouDidNotStartAsMonarch(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	b43Catalog(g, owner.ID, "Knights of the Black Rose", "Creature — Human Knight", knightsOfTheBlackRoseOracle, 4, 4)
	// Seat 0's turn 1 began with nobody wearing the crown.
	monCrown(t, g, owner.ID)
	thiefLife := thief.Life
	monCrown(t, g, thief.ID)
	monSettle(t, g)
	if thief.Life != thiefLife {
		t.Errorf("drained %d on a turn you did not start as the monarch", thiefLife-thief.Life)
	}
}

// --- Entourage of Trest ----------------------------------------------

func TestEntourageOfTrestBlocksAnExtraCreatureOnlyWhileYouAreTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	elf := b43Catalog(g, me.ID, "Entourage of Trest", "Creature — Elf Soldier", entourageOfTrestOracle, 4, 4)
	capacity := func() int {
		c := e2Card(t, g, elf)
		return game.BlockCapacity(&c)
	}
	if got := capacity(); got != 1 {
		t.Fatalf("block capacity without the crown = %d, want 1", got)
	}
	monCrown(t, g, me.ID)
	if got := capacity(); got != 2 {
		t.Errorf("block capacity as the monarch = %d, want 2", got)
	}
	monCrown(t, g, them.ID)
	if got := capacity(); got != 1 {
		t.Errorf("block capacity after losing the crown = %d, want 1 — the static went stale", got)
	}
}

// --- Regal Behemoth --------------------------------------------------

func TestRegalBehemothAddsAManaOnlyWhileYouAreTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	b31Push(g, me.ID, "Regal Behemoth", "Creature — Dinosaur", regalBehemothOracle, "", 5, 5)
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	other := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)

	monCrown(t, g, them.ID)
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 1 {
		t.Fatalf("pool = %v without the crown, want one green", got)
	}

	monCrown(t, g, me.ID)
	tapForMana(t, g, me.ID, other)
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick for the additional mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := sortedPool(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "U" {
		t.Errorf("pool = %v as the monarch, want G G U", got)
	}
}

// --- Throne of the High City -----------------------------------------

func TestThroneOfTheHighCityCrownsYouAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	throne := pushStampedCatalogCard(g, me.ID, "Throne of the High City", "Land", throneOfTheHighCityOracle)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	for i := 0; i < 4; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if z := e2Zone(g, throne); z != game.ZoneGraveyard {
		t.Errorf("the Throne is in %q after paying its cost, want the graveyard", z)
	}
	if g.Monarch == me.ID {
		t.Fatal("the crown moved before the ability resolved")
	}
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want %v", g.Monarch, me.ID)
	}
}

func TestThroneOfTheHighCityCannotBeActivatedWithoutTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	throne := pushStampedCatalogCard(g, me.ID, "Throne of the High City", "Land", throneOfTheHighCityOracle)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {4}, {T}, Sacrifice with an empty pool")
	}
	if g.Monarch != uuid.Nil || e2Zone(g, throne) != game.ZoneBattlefield {
		t.Error("a refused activation moved the crown or the land")
	}
}

// --- Palace Jailer ---------------------------------------------------

// jailerExile casts Palace Jailer for seat 0 and exiles `victim`.
func jailerExile(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	jailer := castAndResolveCreature(t, g, "Palace Jailer", "Creature — Human Soldier", palaceJailerOracle)
	pickCard(t, g, me.ID, victim)
	monSettle(t, g)
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature is in %q, want exile", z)
	}
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want the Jailer's controller", g.Monarch)
	}
	return jailer
}

func TestPalaceJailerReturnsTheCreatureWhenAnOpponentBecomesTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, owner, thief := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailer := jailerExile(t, g, victim)

	// Neither the Jailer leaving, nor the turn ending, nor the crown
	// being cleared, nor YOU becoming the monarch releases it: the
	// duration is keyed to an OPPONENT taking the crown.
	e2Destroy(t, g, jailer)
	advanceToUpkeepOf(t, g, 1)
	monSettle(t, g)
	monCrown(t, g, uuid.Nil)
	monCrown(t, g, me.ID)
	monSettle(t, g)
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature left exile early (in %q)", z)
	}

	monCrown(t, g, thief.ID)
	monSettle(t, g)
	back, ok := e2BattlefieldNamed(g, "Their Bear")
	if !ok {
		t.Fatal("an opponent became the monarch and the creature did not return")
	}
	if back.Controller != owner.ID {
		t.Errorf("returned under %v, want its owner %v", back.Controller, owner.ID)
	}
	// It fires once: the crown moving again returns nothing twice.
	monCrown(t, g, owner.ID)
	monSettle(t, g)
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
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	castAndResolveCreature(t, g, "Palace Jailer", "Creature — Human Soldier", palaceJailerOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if hasID(p.PickTargetCards, mine) {
		t.Error("your own creature is offered")
	}
	if !hasID(p.PickTargetCards, theirs) {
		t.Error("the opponent's creature is not offered")
	}
}

// TestPalaceJailerExileSurvivesASnapshotRoundTrip: the release is a
// keyed delayed trigger, so a table with a creature jailed is a restore
// point, and the restored game still releases it.
func TestPalaceJailerExileSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailerExile(t, g, victim)

	restored := restoreRoundTrip(t, g, true)
	if restored.Monarch != g.Seats[0].ID {
		t.Fatalf("restored monarch = %v, want %v", restored.Monarch, g.Seats[0].ID)
	}
	monCrown(t, restored, thief.ID)
	monSettle(t, restored)
	if _, ok := e2BattlefieldNamed(restored, "Their Bear"); !ok {
		t.Error("the restored game did not release the creature")
	}
}

// TestPalaceJailerControllerLeavingKeepsTheCreatureExiled pins the
// declared caveat: the Jailer's controller concedes while wearing the
// crown, CR 724.4 hands it to an opponent of theirs — which is the
// printed "until" — but the delayed trigger left with its controller
// (CR 800.4a), so the creature stays exiled. Flip this when the gap
// closes, and clear the caveat.
func TestPalaceJailerControllerLeavingKeepsTheCreatureExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailerExile(t, g, victim)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	monSettle(t, g)
	if g.Monarch == uuid.Nil || g.Monarch == me.ID {
		t.Fatalf("CR 724.4: monarch = %v, want a player still in the game", g.Monarch)
	}
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Errorf("the creature is in %q — the caveat has closed; update the card and this test", z)
	}
}

// TestPalaceJailerExileSurvivesUndo: Clone / RestoreFrom keeps the
// delayed trigger and the crown.
func TestPalaceJailerExileSurvivesUndo(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailerExile(t, g, victim)

	before := g.Clone()
	monCrown(t, g, thief.ID)
	monSettle(t, g)
	g.RestoreFrom(before)
	if g.Monarch != g.Seats[0].ID {
		t.Fatalf("after undo monarch = %v, want %v", g.Monarch, g.Seats[0].ID)
	}
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("after undo the creature is in %q, want exile", z)
	}
	monCrown(t, g, thief.ID)
	monSettle(t, g)
	if _, ok := e2BattlefieldNamed(g, "Their Bear"); !ok {
		t.Error("after undo the crown moved and the creature did not return")
	}
}
