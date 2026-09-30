package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_controller_cards_test.go — ADR 0102 PR 2: the three cards that
// enter under the control of an opponent of your choice, end to end
// through a real cast.

const (
	captiveAudienceOracle     = "fc25ca51-f351-4877-a092-525fb48524a1"
	pendantOfProsperityOracle = "d958321b-789d-4f9a-bdbe-f907166a0216"
	abbyMercilessSoldierOracl = "e05efd01-0202-4bd6-bac6-c6210819d463"
)

func tokensNamedControlledBy(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && IsToken(c) && c.Controller == controller {
			n++
		}
	}
	return n
}

// --- Captive Audience ---------------------------------------------------

// The whole card: the caster gives it away, it triggers on the NEW
// controller's upkeep (the 2019 ruling: "that player makes all choices
// for it"), "your life total becomes 4" is theirs, and "each opponent
// creates five Zombies" makes the caster's five.
func TestCaptiveAudienceRunsOnItsNewControllersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victimSeat := (seat + 1) % 4
	victim := g.Seats[victimSeat]

	id := castCatalogSpell(t, g, "Captive Audience", "Enchantment", captiveAudienceOracle, nil)
	passPriorityAroundTable(t, g)
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("Captive Audience did not ask whose control it enters under")
	}
	if c.ControlPurpose != game.ControlForHarm {
		t.Fatalf("purpose = %q, want harm", c.ControlPurpose)
	}
	answerEntryController(t, g, victim.ID)
	card, ok := battlefieldCardByID(g, id)
	if !ok || card.Controller != victim.ID || card.Owner != me.ID {
		t.Fatalf("Captive Audience landed as %+v, want controlled by the victim and owned by the caster", card)
	}

	// Its upkeep trigger is the controller's, not the caster's.
	advanceToUpkeepOf(t, g, victimSeat)
	noModePick(t, g, me.ID, "the caster does not control it")
	chooseModeNow(t, g, victim.ID, []int{0, 1, 2}, 0)
	passPriorityAroundTable(t, g)
	if victim.Life != 4 {
		t.Fatalf("victim life = %d, want 4", victim.Life)
	}

	// Next upkeep: the life mode is used. The Zombies go to the victim's
	// opponents — the caster among them. (The caster's own upkeep in
	// between asks nothing.)
	advanceToUpkeepOf(t, g, seat)
	noModePick(t, g, me.ID, "the caster does not control it")
	advanceToUpkeepOf(t, g, victimSeat)
	chooseModeNow(t, g, victim.ID, []int{1, 2}, 2)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Zombie", me.ID); n != 5 {
		t.Fatalf("the caster has %d Zombies, want 5 (the victim's opponents create them)", n)
	}
	if n := tokensNamedControlledBy(g, "Zombie", victim.ID); n != 0 {
		t.Fatalf("the victim created %d Zombies for itself", n)
	}
}

// --- Pendant of Prosperity ------------------------------------------------

// The controller draws first, then the OWNER — the caster — draws. It
// enters as a gift (benefit), so the purpose says so.
func TestPendantOfProsperityDrawsForItsControllerThenItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	friendSeat := (seat + 1) % 4
	friend := g.Seats[friendSeat]

	id := castCatalogSpell(t, g, "Pendant of Prosperity", "Artifact", pendantOfProsperityOracle, nil)
	passPriorityAroundTable(t, g)
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("Pendant of Prosperity did not ask whose control it enters under")
	}
	if c.ControlPurpose != game.ControlForBenefit {
		t.Fatalf("purpose = %q, want benefit", c.ControlPurpose)
	}
	answerEntryController(t, g, friend.ID)

	advanceToUpkeepOf(t, g, friendSeat)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.AddManaForEffect(friend.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	// A land in the controller's hand, so their "you may put a land"
	// is a real question the owner's half has to wait for.
	g.WithWriteLock(func() {
		friend.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: friend.ID, Controller: friend.ID})
	})
	friendHand, myHand := handSize(friend), handSize(me)
	if err := g.ActivateCatalogAbility(friend.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Pendant: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The controller's land prompt comes first; the owner has not drawn
	// yet.
	if handSize(me) != myHand {
		t.Fatal("the owner drew before the controller answered their land prompt")
	}
	answerLandPrompt(t, g, friend.ID)
	answerLandPrompt(t, g, me.ID)
	if handSize(friend) != friendHand+1 {
		t.Fatalf("controller hand %d → %d, want +1", friendHand, handSize(friend))
	}
	if handSize(me) != myHand+1 {
		t.Fatalf("owner hand %d → %d, want +1", myHand, handSize(me))
	}
}

// answerLandPrompt declines a "you may put a land card" prompt owed by
// `chooser`, if there is one.
func answerLandPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == chooser && c.Kind == game.PendingChoiceChooseCards {
			if err := g.ResolveChooseCards(c.ID, chooser, nil); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			return
		}
	}
}

// --- Abby, Merciless Soldier ----------------------------------------------

// The cast trigger is the caster's and counts the mana spent; Abby
// herself enters under an opponent.
func TestAbbyMakesTheCasterTokensAndEntersUnderAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	them := g.Seats[(seat+2)%4]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{R}{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	abby := castFromHandForTest(t, g, me, "Abby, Merciless Soldier", "Legendary Creature — Human Survivor",
		"{1}{R}{G}", abbyMercilessSoldierOracl, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Cordyceps Infected", me.ID); n != 3 {
		t.Fatalf("the caster has %d Cordyceps Infected, want 3 (one per mana spent)", n)
	}
	answerEntryController(t, g, them.ID)
	card, ok := battlefieldCardByID(g, abby)
	if !ok || card.Controller != them.ID || card.Owner != me.ID {
		t.Fatalf("Abby landed as %+v, want controlled by the chosen opponent", card)
	}
}

// A countered Abby still leaves the caster the tokens: the trigger is a
// separate object on the stack above her.
func TestAbbyCounteredStillMakesTheTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{R}{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	abby := castFromHandForTest(t, g, me, "Abby, Merciless Soldier", "Legendary Creature — Human Survivor",
		"{1}{R}{G}", abbyMercilessSoldierOracl, game.CastSpellParams{Strict: true})
	if still := counterByEffect(t, g, abby); still {
		t.Fatal("Abby was not countered")
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Cordyceps Infected", me.ID); n != 3 {
		t.Fatalf("the caster has %d Cordyceps Infected after a counter, want 3", n)
	}
	if entryControllerPrompt(g) != nil {
		t.Fatal("a countered Abby asked whose control it enters under")
	}
}
