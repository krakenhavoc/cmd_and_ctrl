package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b19_hideaway_test.go — Wiretapping, Widespread Thieving and
// Smuggler's Buggy: the linked abilities over the card each one hid.

const (
	wiretappingOracle       = "befc5986-f06c-40a4-98ee-ffcace4ceaa8"
	widespreadThievingOracl = "2c2e4b19-ac88-426a-9f1b-ac305cd512e8"
	smugglersBuggyOracle    = "e19e1e83-75e5-4cfa-8d50-8ed0b027a202"
)

// sncB19Hide puts a hideaway permanent on p's battlefield and hides a fresh
// card with it, as Hideaway's resolution does.
func sncB19Hide(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, hidden game.Card) (perm, hid uuid.UUID) {
	t.Helper()
	perm = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 5, Toughness: 5, Owner: p.ID, Controller: p.ID,
	})
	hidden.InstanceID = uuid.New()
	hidden.Owner = p.ID
	p.Library.PushTop(hidden)
	g.WithWriteLock(func() {
		ref := game.ObjectRefOf(*findBattlefieldCardForTest(g, perm))
		if paused, err := g.ExileHiddenForEffect(ref, p.ID, hidden.InstanceID); paused || err != nil {
			t.Fatalf("ExileHiddenForEffect = %v, %v", paused, err)
		}
	})
	return perm, hidden.InstanceID
}

var sncB19HiddenSpell = game.Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"}

func TestSncB19WiretappingDrawsOnTheFirstDrawStepDrawAndPlaysAtNineCards(t *testing.T) {
	for _, tc := range []struct {
		name      string
		extra     int // cards in hand before the turn draw
		wantGrant bool
	}{
		{"eight cards after both draws", 6, false},
		{"nine cards after both draws", 7, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			_, hid := sncB19Hide(t, g, me, "Wiretapping", "Enchantment", wiretappingOracle, sncB19HiddenSpell)
			for me.Hand.Size() > tc.extra {
				if _, err := me.Hand.PopTop(); err != nil {
					t.Fatal(err)
				}
			}
			for me.Hand.Size() < tc.extra {
				me.Hand.PushTop(game.NewCard("Filler", me.ID))
			}
			for i := 0; i < 4; i++ {
				me.Library.PushTop(game.NewCard("Lib", me.ID))
			}
			advanceToDrawStepOfSeat(t, g, 1)
			handAtStep := me.Hand.Size()
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != handAtStep+1 {
				t.Fatalf("hand = %d, want one extra card (%d)", got, handAtStep+1)
			}
			if got := playHiddenGrant(g, hid) != nil; got != tc.wantGrant {
				t.Fatalf("grant over the hidden card = %v, want %v (hand %d)", got, tc.wantGrant, me.Hand.Size())
			}
			// The extra draw is the second of the step and does not trigger again.
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != handAtStep+1 {
				t.Fatalf("hand = %d after settling, want %d: the extra draw retriggered", got, handAtStep+1)
			}
		})
	}
}

func TestSncB19WiretappingIgnoresAnOpponentsDrawStepAndLaterDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sncB19Hide(t, g, me, "Wiretapping", "Enchantment", wiretappingOracle, sncB19HiddenSpell)
	advanceToDrawStepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 {
		t.Fatal("Wiretapping triggered on an opponent's draw")
	}
	drawFor(t, g, g.Seats[1], 1)
	if g.Stack.Size() != 0 {
		t.Fatal("Wiretapping triggered on an opponent's draw")
	}
	// Its controller's own second draw of a draw step is not the first card.
	advanceToDrawStepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	drawFor(t, g, me, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Fatalf("a later draw triggered Wiretapping: hand %d -> %d", before, me.Hand.Size())
	}
}

func TestSncB19WidespreadThievingTreasureThenPaysFiveColorsForTheHiddenCard(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		_, hid := sncB19Hide(t, g, me, "Widespread Thieving", "Enchantment", widespreadThievingOracl, sncB19HiddenSpell)
		spell := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: spell, Name: "Gold Spell", TypeLine: "Sorcery", ManaCost: "{R}{G}",
			Colors: []string{"R", "G"}, Owner: me.ID, Controller: me.ID,
		})
		floatForTest(g, me, "RG")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g) // trigger resolves: Treasure, then the payment is offered
		treasures := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.Controller == me.ID {
				treasures++
			}
		}
		if treasures != 1 {
			t.Fatalf("Treasures = %d, want 1", treasures)
		}
		if payc := latestChoiceOfKind(g, game.PendingChoicePayUnless); payc == nil {
			t.Fatal("the {W}{U}{B}{R}{G} payment was not offered")
		} else {
			if pay {
				floatForTest(g, me, "WUBRG")
			}
			if err := g.ResolvePayUnless(payc.ID, me.ID, pay); err != nil {
				t.Fatalf("ResolvePayUnless(%v): %v", pay, err)
			}
		}
		if got := playHiddenGrant(g, hid) != nil; got != pay {
			t.Errorf("pay=%v: grant over the hidden card = %v", pay, got)
		}
	}
}

func TestSncB19WidespreadThievingIgnoresMonocoloredAndOpposingSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	sncB19Hide(t, g, me, "Widespread Thieving", "Enchantment", widespreadThievingOracl, sncB19HiddenSpell)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Mono Spell", TypeLine: "Sorcery", ManaCost: "{R}",
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" {
			t.Fatal("a monocolored spell made a Treasure")
		}
	}
}

// sncB19BuggyHit hides `hidden` under a Buggy and has it connect.
func sncB19BuggyHit(t *testing.T, hidden game.Card) (g *game.Game, me *game.Player, buggy, hid uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	buggy, hid = sncB19Hide(t, g, me, "Smuggler's Buggy", "Artifact — Vehicle", smugglersBuggyOracle, hidden)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, buggy)
		c.SummonedThisTurn = false
		c.Colors = nil
	})
	// A Vehicle is only a creature once crewed: make it one for the swing.
	crewer := pushVanillaCreature(g, me.ID, "Crew", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	crewBy(t, g, me.ID, buggy, crewer)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(buggy, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	return
}

func TestSncB19SmugglersBuggyCastingTheHiddenCardReturnsTheVehicle(t *testing.T) {
	g, me, buggy, hid := sncB19BuggyHit(t, sncB19HiddenSpell)
	answerMayCastPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	t.Logf("after answer: stack=%d pending=%d buggyOnBF=%v inHand=%v inExile=%v", g.Stack.Size(), len(g.PendingChoices), g.Battlefield.Contains(buggy), me.Hand.Contains(buggy), inExile(g, hid))
	if !me.Hand.Contains(buggy) && g.Battlefield.Contains(buggy) {
		t.Fatalf("Buggy stayed on the battlefield after the hidden card was cast")
	}
	if inExile(g, hid) {
		t.Fatal("the hidden card is still in exile after being cast")
	}
}

func TestSncB19SmugglersBuggyDecliningLeavesBothWhereTheyAre(t *testing.T) {
	g, me, buggy, hid := sncB19BuggyHit(t, sncB19HiddenSpell)
	answerMayCastPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(buggy) {
		t.Fatal("declining returned the Vehicle")
	}
	if !inExile(g, hid) {
		t.Fatal("declining moved the hidden card")
	}
}

func TestSncB19SmugglersBuggyHiddenLandIsNotCastAndTheVehicleStays(t *testing.T) {
	g, _, buggy, _ := sncB19BuggyHit(t, game.Card{Name: "Hidden Forest", TypeLine: "Basic Land — Forest"})
	if !g.Battlefield.Contains(buggy) {
		t.Fatal("a land can't be cast, so nothing was cast and the Vehicle must stay")
	}
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) != nil {
		t.Fatal("a cast was offered over a land")
	}
}
