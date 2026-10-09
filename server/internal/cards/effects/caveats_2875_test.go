package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	c2875CultivateOracle = "8b755881-a72d-4e21-a369-d2924eb4585a"
	c2875KodamaOracle    = "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90"
	c2875UprisingOracle  = "3127ae9b-a7a7-43ec-89d7-688f8445b33d"
)

// stackHasTrigger reports whether any triggered ability is on the stack.
func stackHasTrigger(g *game.Game) bool {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered {
			return true
		}
	}
	return false
}

// #2875: Berserk's window is per combat. After an earlier combat's
// damage step, the next combat's steps before ITS damage step are open
// again, and its own damage step is closed.
func TestBerserkIsCastablePerCombatInATurnWithTwoCombats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aurelia := pushDiesCreatureForTest(g, me.ID, "Aurelia, the Warleader", aureliaWarleaderOracle, "Legendary Creature — Angel", 3, 4)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aurelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	if err := s58p6CastNow(g, me, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(aurelia)); err != nil {
		t.Fatalf("beginning of the second combat, after the first combat's damage: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	if err := s58p6CastNow(g, me, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(aurelia)); err == nil {
		t.Error("cast in the second combat's damage step")
	}
}

// #2875: Cultivate with exactly one basic in the library asks where it
// goes; either answer is honoured, and the library is shuffled once.
func TestCultivateWithOneBasicLetsYouChooseHandOrBattlefield(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, card string
		toBattlefield      bool
	}{
		{"Cultivate battlefield", c2875CultivateOracle, "Cultivate", true},
		{"Cultivate hand", c2875CultivateOracle, "Cultivate", false},
		{"Kodama's Reach battlefield", c2875KodamaOracle, "Kodama's Reach", true},
		{"Kodama's Reach hand", c2875KodamaOracle, "Kodama's Reach", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Library.Cards = nil
			forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			pushLibraryCardForTest(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"})

			castCatalogSpell(t, g, tc.card, "Sorcery", tc.oracle, nil)
			passPriorityAroundTable(t, g)
			ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
			if ask == nil {
				t.Fatal("no battlefield-or-hand question for a lone basic land")
			}
			if err := g.ResolveConfirm(ask.ID, me.ID, tc.toBattlefield); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			if got := g.Battlefield.Contains(forest); got != tc.toBattlefield {
				t.Errorf("on the battlefield: got %v, want %v", got, tc.toBattlefield)
			}
			if got := me.Hand.Contains(forest); got == tc.toBattlefield {
				t.Errorf("in hand: got %v, want %v", got, !tc.toBattlefield)
			}
			if tc.toBattlefield && !isTapped(g, forest) {
				t.Error("the Forest entered untapped")
			}
		})
	}
}

// #2875: Garruk's Uprising's enters draw is an intervening if (CR
// 603.4): the power-4 creature removed in response leaves no card.
func TestGarruksUprisingDrawIsRecheckedOnResolution(t *testing.T) {
	for _, tc := range []struct {
		name    string
		removed bool
		drawn   int
	}{
		{"creature stays", false, 1},
		{"creature removed in response", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			big := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
				Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
			})
			castCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", c2875UprisingOracle, nil)
			// Pass until the Uprising resolves and its enters trigger
			// is on the stack, then stop: the response comes here.
			for i := 0; i < 8 && !stackHasTrigger(g); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if !stackHasTrigger(g) {
				t.Fatal("the enters trigger is not on the stack")
			}
			before := me.Hand.Size()
			if tc.removed {
				killCreature(t, g, me.ID, big)
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.drawn {
				t.Errorf("cards drawn: got %d, want %d", got, tc.drawn)
			}
		})
	}
}
