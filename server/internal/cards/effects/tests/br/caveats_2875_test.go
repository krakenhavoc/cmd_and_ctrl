package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aurelia := effectstest.PushDiesCreatureForTest(g, me.ID, "Aurelia, the Warleader", effectstest.AureliaWarleaderOracle, "Legendary Creature — Angel", 3, 4)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aurelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	if err := effectstest.S58p6CastNow(g, me, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(aurelia)); err != nil {
		t.Fatalf("beginning of the second combat, after the first combat's damage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The additional combat needs an attacker of its own to have a
	// combat damage step at all (CR 508.8).
	runner := effectstest.PushVanillaCreature(g, me.ID, "Runner", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if err := effectstest.S58p6CastNow(g, me, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(aurelia)); err == nil {
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
		{"Cultivate battlefield", effectstest.C2875CultivateOracle, "Cultivate", true},
		{"Cultivate hand", effectstest.C2875CultivateOracle, "Cultivate", false},
		{"Kodama's Reach battlefield", effectstest.C2875KodamaOracle, "Kodama's Reach", true},
		{"Kodama's Reach hand", effectstest.C2875KodamaOracle, "Kodama's Reach", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Library.Cards = nil
			forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"})

			effectstest.CastCatalogSpell(t, g, tc.card, "Sorcery", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			// #2942: the lone basic is offered for the battlefield
			// first and, if declined, for the hand.
			if tc.toBattlefield {
				effectstest.AnswerSearchByID(t, g, me.ID, forest)
			} else {
				effectstest.AnswerSearchFailToFind(t, g, me.ID)
				effectstest.AnswerSearchByID(t, g, me.ID, forest)
			}
			if got := g.Battlefield.Contains(forest); got != tc.toBattlefield {
				t.Errorf("on the battlefield: got %v, want %v", got, tc.toBattlefield)
			}
			if got := me.Hand.Contains(forest); got == tc.toBattlefield {
				t.Errorf("in hand: got %v, want %v", got, !tc.toBattlefield)
			}
			if tc.toBattlefield && !effectstest.IsTapped(g, forest) {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
				Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
			})
			effectstest.CastCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", effectstest.C2875UprisingOracle, nil)
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
				effectstest.KillCreature(t, g, me.ID, big)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.drawn {
				t.Errorf("cards drawn: got %d, want %d", got, tc.drawn)
			}
		})
	}
}
