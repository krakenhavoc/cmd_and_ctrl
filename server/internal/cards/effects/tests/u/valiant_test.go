package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestValiantFiresTheFirstTimeYouTargetItThisTurn(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	if n := effectstest.ValiantTriggersFrom(tb.G, tb.Hero); n != 1 {
		t.Fatalf("%d valiant triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters, want 1", n)
	}
}

func TestValiantDoesNotFireASecondTimeInTheSameTurn(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	effectstest.PassPriorityAroundTable(t, tb.G)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	if n := effectstest.ValiantTriggersFrom(tb.G, tb.Hero); n != 0 {
		t.Fatalf("the second spell this turn put %d valiant triggers on the stack", n)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters after two spells in one turn, want 1", n)
	}
}

func TestValiantFiresAgainOnTheNextTurn(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	effectstest.PassPriorityAroundTable(t, tb.G)
	effectstest.ToNextTurnMainOf(t, tb.G, (tb.Seat+1)%len(tb.G.Seats))
	// An opponent's turn is still "each turn": the controller's instant then
	// is the first time this turn.
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 2 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters after one spell on each of two turns, want 2", n)
	}
}

func TestValiantIgnoresAnOpponentsSpell(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Them, tb.Hero)
	if n := effectstest.ValiantTriggersFrom(tb.G, tb.Hero); n != 0 {
		t.Fatalf("an opponent's spell put %d valiant triggers on the stack", n)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	// The opponent's targeting does not use up the turn's first time.
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters, want 1 (from its controller's spell only)", n)
	}
}

func TestValiantFiresOnceForOneSpellTargetingItInTwoSlots(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero, tb.Hero)
	if n := effectstest.ValiantTriggersFrom(tb.G, tb.Hero); n != 1 {
		t.Fatalf("one spell with two target slots on Heartfire Hero put %d valiant triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters, want 1", n)
	}
}

// CR 400.7: a Hero that left and came back is a new object that has not
// been targeted this turn. A bounce and a recast keep the card's instance
// ID (only an exile return mints a new one), so this is the case the
// zone-change reset in the scan is for.
func TestValiantFiresAgainForANewObjectTheSameTurn(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	effectstest.PassPriorityAroundTable(t, tb.G)
	tb.G.WithWriteLock(func() {
		if err := tb.G.BounceToHandForEffect(tb.Hero); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if !tb.Me.Hand.Contains(tb.Hero) {
		t.Fatal("Heartfire Hero is not in its owner's hand under the same ID")
	}
	if err := tb.G.CastSpell(tb.Me.ID, tb.Hero, game.CastSpellParams{}); err != nil {
		t.Fatalf("recast Heartfire Hero: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if !effectstest.OnBoard(tb.G, tb.Hero) {
		t.Fatal("Heartfire Hero did not come back under the same ID")
	}
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 0 {
		t.Fatalf("the returned Hero kept %d counters", n)
	}
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, tb.Hero); n != 1 {
		t.Fatalf("the returned Hero has %d +1/+1 counters after its first targeting, want 1", n)
	}
}

// The trigger of a targeted valiant ability waits on its target prompt, so
// it has not been announced when the spell's second slot is: the gate must
// still say "already targeted".
func TestValiantTargetedAbilityFiresOnceForTwoSlots(t *testing.T) {
	tb, trapper := effectstest.NewMouseTrapperTable(t)
	// Two candidates, so the target is a real prompt and not an automatic pick.
	bear := effectstest.B12Creature(tb.G, tb.Them.ID, "Opposing Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(tb.G, tb.Them.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.ValiantTarget(t, tb.G, tb.Me, trapper, trapper)
	if n := effectstest.ValiantPickTargetPrompts(tb.G) + effectstest.ValiantTriggersFrom(tb.G, trapper); n != 1 {
		t.Fatalf("one two-slot spell made %d Mouse Trapper triggers (prompts plus items), want 1", n)
	}
	effectstest.AnswerPickTarget(t, tb.G, bear)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if !effectstest.TappedNow(tb.G, bear) {
		t.Error("Mouse Trapper did not tap the opposing creature")
	}
	if effectstest.TappedNow(tb.G, other) || effectstest.TappedNow(tb.G, trapper) {
		t.Error("Mouse Trapper tapped something it did not target")
	}
}

// CR 603.3d: with no creature an opponent controls, the trigger is removed,
// but the Trapper has been targeted, so a later spell that turn does not
// trigger it either.
func TestValiantRemovedForNoTargetStillUsesTheTurn(t *testing.T) {
	tb, trapper := effectstest.NewMouseTrapperTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, trapper)
	if n := effectstest.ValiantPickTargetPrompts(tb.G) + effectstest.ValiantTriggersFrom(tb.G, trapper); n != 0 {
		t.Fatalf("Mouse Trapper triggered %d times with nothing to tap", n)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	effectstest.B12Creature(tb.G, tb.Them.ID, "Late Bear", "Creature — Bear", 2, 2)
	effectstest.ValiantTarget(t, tb.G, tb.Me, trapper)
	if n := effectstest.ValiantPickTargetPrompts(tb.G) + effectstest.ValiantTriggersFrom(tb.G, trapper); n != 0 {
		t.Fatalf("Mouse Trapper triggered %d times on the second targeting of the turn", n)
	}
}
