package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	perm = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 5, Toughness: 5, Owner: p.ID, Controller: p.ID,
	})
	hidden.InstanceID = uuid.New()
	hidden.Owner = p.ID
	p.Library.PushTop(hidden)
	g.WithWriteLock(func() {
		ref := game.ObjectRefOf(*effectstest.FindBattlefieldCardForTest(g, perm))
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
			g := effectstest.NewCatalogGame(t)
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
			effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
			handAtStep := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != handAtStep+1 {
				t.Fatalf("hand = %d, want one extra card (%d)", got, handAtStep+1)
			}
			if got := effectstest.PlayHiddenGrant(g, hid) != nil; got != tc.wantGrant {
				t.Fatalf("grant over the hidden card = %v, want %v (hand %d)", got, tc.wantGrant, me.Hand.Size())
			}
			// The extra draw is the second of the step and does not trigger again.
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != handAtStep+1 {
				t.Fatalf("hand = %d after settling, want %d: the extra draw retriggered", got, handAtStep+1)
			}
		})
	}
}

func TestSncB19WiretappingIgnoresAnOpponentsDrawStepAndLaterDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sncB19Hide(t, g, me, "Wiretapping", "Enchantment", wiretappingOracle, sncB19HiddenSpell)
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 {
		t.Fatal("Wiretapping triggered on an opponent's draw")
	}
	effectstest.DrawFor(t, g, g.Seats[1], 1)
	if g.Stack.Size() != 0 {
		t.Fatal("Wiretapping triggered on an opponent's draw")
	}
	// Its controller's own second draw of a draw step is not the first card.
	effectstest.AdvanceToDrawStepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()
	effectstest.DrawFor(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Fatalf("a later draw triggered Wiretapping: hand %d -> %d", before, me.Hand.Size())
	}
}

func TestSncB19WidespreadThievingTreasureThenPaysFiveColorsForTheHiddenCard(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		_, hid := sncB19Hide(t, g, me, "Widespread Thieving", "Enchantment", widespreadThievingOracl, sncB19HiddenSpell)
		spell := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: spell, Name: "Gold Spell", TypeLine: "Sorcery", ManaCost: "{R}{G}",
			Colors: []string{"R", "G"}, Owner: me.ID, Controller: me.ID,
		})
		effectstest.FloatForTest(g, me, "RG")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g) // trigger resolves: Treasure, then the payment is offered
		treasures := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.Controller == me.ID {
				treasures++
			}
		}
		if treasures != 1 {
			t.Fatalf("Treasures = %d, want 1", treasures)
		}
		if payc := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless); payc == nil {
			t.Fatal("the {W}{U}{B}{R}{G} payment was not offered")
		} else {
			if pay {
				effectstest.FloatForTest(g, me, "WUBRG")
			}
			if err := g.ResolvePayUnless(payc.ID, me.ID, pay); err != nil {
				t.Fatalf("ResolvePayUnless(%v): %v", pay, err)
			}
		}
		if got := effectstest.PlayHiddenGrant(g, hid) != nil; got != pay {
			t.Errorf("pay=%v: grant over the hidden card = %v", pay, got)
		}
	}
}

func TestSncB19WidespreadThievingIgnoresMonocoloredAndOpposingSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	sncB19Hide(t, g, me, "Widespread Thieving", "Enchantment", widespreadThievingOracl, sncB19HiddenSpell)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Mono Spell", TypeLine: "Sorcery", ManaCost: "{R}",
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" {
			t.Fatal("a monocolored spell made a Treasure")
		}
	}
}

// sncB19BuggyHit hides `hidden` under a Buggy and has it connect.
func sncB19BuggyHit(t *testing.T, hidden game.Card) (g *game.Game, me *game.Player, buggy, hid uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	buggy, hid = sncB19Hide(t, g, me, "Smuggler's Buggy", "Artifact — Vehicle", smugglersBuggyOracle, hidden)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, buggy)
		c.SummonedThisTurn = false
		c.Colors = nil
	})
	// A Vehicle is only a creature once crewed: make it one for the swing.
	crewer := effectstest.PushVanillaCreature(g, me.ID, "Crew", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.CrewBy(t, g, me.ID, buggy, crewer)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(buggy, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestSncB19SmugglersBuggyCastingTheHiddenCardReturnsTheVehicle(t *testing.T) {
	g, me, buggy, hid := sncB19BuggyHit(t, sncB19HiddenSpell)
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	t.Logf("after answer: stack=%d pending=%d buggyOnBF=%v inHand=%v inExile=%v", g.Stack.Size(), len(g.PendingChoices), g.Battlefield.Contains(buggy), me.Hand.Contains(buggy), effectstest.InExile(g, hid))
	if !me.Hand.Contains(buggy) && g.Battlefield.Contains(buggy) {
		t.Fatalf("Buggy stayed on the battlefield after the hidden card was cast")
	}
	if effectstest.InExile(g, hid) {
		t.Fatal("the hidden card is still in exile after being cast")
	}
}

func TestSncB19SmugglersBuggyDecliningLeavesBothWhereTheyAre(t *testing.T) {
	g, me, buggy, hid := sncB19BuggyHit(t, sncB19HiddenSpell)
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(buggy) {
		t.Fatal("declining returned the Vehicle")
	}
	if !effectstest.InExile(g, hid) {
		t.Fatal("declining moved the hidden card")
	}
}

func TestSncB19SmugglersBuggyHiddenLandIsNotCastAndTheVehicleStays(t *testing.T) {
	g, _, buggy, _ := sncB19BuggyHit(t, game.Card{Name: "Hidden Forest", TypeLine: "Basic Land — Forest"})
	if !g.Battlefield.Contains(buggy) {
		t.Fatal("a land can't be cast, so nothing was cast and the Vehicle must stay")
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast) != nil {
		t.Fatal("a cast was offered over a land")
	}
}
