package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// karplusan_minotaur_test.go — #4063: "whenever you lose a coin flip" and a
// cumulative upkeep paid by flipping a coin, on Karplusan Minotaur.

const karplusanMinotaurOracle = "b3122a15-c7ab-40f1-9632-f5fbc4484c59"

// minotaurWithAge puts a Karplusan Minotaur onto the battlefield under
// owner with `age` age counters already on it.
func minotaurWithAge(g *game.Game, owner uuid.UUID, age int) uuid.UUID {
	id := pushPermanentForTest(g, owner, "Karplusan Minotaur", karplusanMinotaurOracle, "Creature — Minotaur Warrior")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
			if age > 0 {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterAge: age}
			}
		}
	}
	return id
}

// flipEventsBy counts the called coin flips a player has made so far.
func flipEventsBy(g *game.Game, flipper uuid.UUID) (won, lost int) {
	for _, ev := range g.Events {
		if ev.Kind == game.EventFlipCoin && ev.Actor == flipper && ev.Call != "" {
			if ev.Won {
				won++
			} else {
				lost++
			}
		}
	}
	return
}

// The matcher: only the controller's own called, lost flip.
func TestKarplusanMinotaurMatcherIsTheControllersLostCalledFlipOnly(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	source := &game.Card{Controller: me}
	for _, tc := range []struct {
		name string
		ev   game.Event
		want bool
	}{
		{"my lost flip", game.Event{Kind: game.EventFlipCoin, Actor: me, Call: "heads"}, true},
		{"my won flip", game.Event{Kind: game.EventFlipCoin, Actor: me, Call: "heads", Won: true}, false},
		{"an opponent's lost flip", game.Event{Kind: game.EventFlipCoin, Actor: opp, Call: "heads"}, false},
		{"my face-only flip", game.Event{Kind: game.EventFlipCoin, Actor: me, Label: "tails"}, false},
		{"a table roll", game.Event{Kind: game.EventTableRoll, Actor: me, Call: "heads"}, false},
		{"no flipper", game.Event{Kind: game.EventFlipCoin, Call: "heads"}, false},
	} {
		if got := YouLostACoinFlip(tc.ev, source, game.Characteristic{}, nil); got != tc.want {
			t.Errorf("%s: matches = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A won flip: the controller names the target of the 1 damage.
func TestKarplusanMinotaurWonFlipDealsOneToTargetYouChoose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minotaur := minotaurWithAge(g, me.ID, 0)

	emitCoinFlip(g, me.ID, true)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("a won flip asked the controller for no target")
	}
	before := opp.Life
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-1)
	}
	if item := triggerOnStack(g, minotaur); item != nil {
		t.Errorf("a trigger is still on the stack: %+v", item)
	}
}

// A lost flip: the controller names an opponent, and that opponent names
// the target. The controller cannot aim it.
func TestKarplusanMinotaurLostFlipTargetIsAnOpponentsChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	minotaurWithAge(g, me.ID, 0)

	emitCoinFlip(g, me.ID, false)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the controller was asked for a lost flip's target")
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	otPickFor(t, g, a.ID, me.ID)
	beforeMe, beforeB := me.Life, b.Life
	// Opponent a aims it at the controller's own face.
	otAnswer(t, g, a.ID, otPlayer(me.ID))
	passPriorityAroundTable(t, g)
	if me.Life != beforeMe-1 {
		t.Errorf("controller life %d, want %d: the opponent's pick was not used", me.Life, beforeMe-1)
	}
	if b.Life != beforeB {
		t.Errorf("another opponent's life moved to %d", b.Life)
	}
}

// Neither ability answers a flip that is not the controller's called one:
// an opponent's flips and a face-only flip of the controller's.
func TestKarplusanMinotaurIgnoresOtherFlips(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minotaur := minotaurWithAge(g, me.ID, 0)

	emitCoinFlip(g, opp.ID, true)
	emitCoinFlip(g, opp.ID, false)
	g.WithWriteLock(func() {
		if _, err := g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID}, 3); err != nil {
			t.Fatalf("FlipCoinsForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if triggerOnStack(g, minotaur) != nil || latestPickTarget(g, me.ID) != nil || len(g.PendingChoices) != 0 {
		t.Errorf("an unrelated flip started an ability: stack %v, choices %d", triggerOnStack(g, minotaur), len(g.PendingChoices))
	}
}

// Each coin is its own flip: two lost coins are two triggers.
func TestKarplusanMinotaurFiresOncePerCoin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	minotaur := minotaurWithAge(g, me.ID, 0)
	emitCoinFlip(g, me.ID, true)
	emitCoinFlip(g, me.ID, true)
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == minotaur {
			n++
		}
	}
	// Each trigger may still be parked behind its target pick.
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == me.ID {
			picks++
		}
	}
	if n+picks != 2 {
		t.Errorf("%d triggers on the stack and %d target picks, want two in all", n, picks)
	}
}

// The cumulative upkeep: saying yes flips one coin for the first age
// counter, as one called flip that cannot be stopped, and the flip is
// a won or a lost one for the Minotaur's own abilities.
func TestKarplusanMinotaurCumulativeUpkeepFlipsAndCannotStop(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	minotaur := minotaurWithAge(g, owner.ID, 0)

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := ageCountersOn(t, g, minotaur); got != 1 {
		t.Fatalf("age counters = %d, want 1", got)
	}
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	if a := choice.PayAction(); a == nil || a.Kind != game.PayActionNothing {
		t.Fatalf("the upkeep payment is %+v, want a payment that costs nothing", a)
	}
	if choice.PayCost != "Flip a coin" {
		t.Errorf("the prompt names %q, want %q", choice.PayCost, "Flip a coin")
	}
	if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
		t.Fatalf("say yes to the flip: %v", err)
	}
	if _, ok := battlefieldCard(g, minotaur); !ok {
		t.Fatal("the Minotaur was sacrificed after the controller agreed to flip")
	}
	flip := latestChoiceOfKindFor(g, game.PendingChoiceCoinCall, owner.ID)
	if flip == nil {
		t.Fatal("no coin call after paying")
	}
	if flip.CoinCount != 1 || flip.CoinAllowStop {
		t.Errorf("coin call: %d coins, stoppable %v; want 1 coin and no stopping", flip.CoinCount, flip.CoinAllowStop)
	}
	if err := g.ResolveCoinCall(flip.ID, owner.ID, "stop"); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("stopping the flip: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveCoinCall(flip.ID, owner.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	won, lost := flipEventsBy(g, owner.ID)
	if won+lost != 1 {
		t.Fatalf("%d called flips, want 1", won+lost)
	}
	// The one flip started exactly the matching ability.
	if won == 1 && latestPickTarget(g, owner.ID) == nil {
		t.Error("a won flip asked for no target")
	}
	if lost == 1 && latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, owner.ID) == nil {
		t.Error("a lost flip asked for no opponent")
	}
}

// With two age counters the price is two coins under one call, each its own
// won or lost flip.
func TestKarplusanMinotaurCumulativeUpkeepFlipsOnePerAgeCounter(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	minotaur := minotaurWithAge(g, owner.ID, 1)

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := ageCountersOn(t, g, minotaur); got != 2 {
		t.Fatalf("age counters = %d, want 2", got)
	}
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	if choice.PayCost != "Flip two coins" {
		t.Errorf("the prompt names %q, want %q", choice.PayCost, "Flip two coins")
	}
	if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	flip := latestChoiceOfKindFor(g, game.PendingChoiceCoinCall, owner.ID)
	if flip == nil || flip.CoinCount != 2 {
		t.Fatalf("coin call = %+v, want two coins", flip)
	}
	if err := g.ResolveCoinCall(flip.ID, owner.ID, "tails"); err != nil {
		t.Fatal(err)
	}
	if won, lost := flipEventsBy(g, owner.ID); won+lost != 2 {
		t.Errorf("%d called flips, want 2", won+lost)
	}
}

// Saying no sacrifices the Minotaur and flips nothing.
func TestKarplusanMinotaurCumulativeUpkeepDeclineSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	minotaur := minotaurWithAge(g, owner.ID, 0)

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := battlefieldCard(g, minotaur); ok {
		t.Error("the Minotaur survived a declined cumulative upkeep")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceCoinCall, owner.ID) != nil {
		t.Error("a declined upkeep still asked for a coin call")
	}
}

// The upkeep prompt survives a clone, and the flip is queued on the copy.
func TestKarplusanMinotaurUpkeepPromptSurvivesAClone(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	minotaurWithAge(g, owner.ID, 0)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	choice := remoraUpkeepPrompt(t, g, owner.ID)

	clone := g.Clone()
	if err := clone.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
		t.Fatalf("answering the clone: %v", err)
	}
	if latestChoiceOfKindFor(clone, game.PendingChoiceCoinCall, owner.ID) == nil {
		t.Error("the clone's payment queued no coin call")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceCoinCall, owner.ID) != nil {
		t.Error("answering the clone queued a coin call on the original")
	}
}

// settleMinotaurTrigger answers whichever damage trigger the last flip started
// (a won flip's target, or a lost flip's opponent and that opponent's pick),
// aiming it at the first opponent, and settles the stack.
func settleMinotaurTrigger(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	target := g.Seats[0].ID
	if p := latestPickTarget(g, owner); p != nil {
		if err := g.ResolvePickTargets(p.ID, owner, []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}); err != nil {
			t.Fatalf("won flip target: %v", err)
		}
	} else if c := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, owner); c != nil {
		if err := g.ResolveOptionPick(c.ID, owner, 0); err != nil {
			t.Fatalf("lost flip opponent: %v", err)
		}
		chooser := c.PickOptions[0].Player
		p := latestPickTarget(g, chooser)
		if p == nil {
			t.Fatal("the chosen opponent was asked for no target")
		}
		if err := g.ResolvePickTargets(p.ID, chooser, []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}); err != nil {
			t.Fatalf("lost flip target: %v", err)
		}
	} else {
		t.Fatal("the flip started no ability")
	}
	passPriorityAroundTable(t, g)
}

// The coin is the seeded stream of the game. A table with the coin call open
// is not a restore point (the call holds a continuation), so the flip is
// answered first; a restore taken after it then draws the next flip exactly
// as the original does, and the Minotaur answers it the same way.
func TestKarplusanMinotaurFlipIsReproducibleAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	minotaur := minotaurWithAge(g, owner.ID, 0)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	flip := latestChoiceOfKindFor(g, game.PendingChoiceCoinCall, owner.ID)
	if flip == nil {
		t.Fatal("no coin call after paying")
	}
	if g.CaptureSnapshot().Restorable() {
		t.Error("a table with an open coin call is a restore point")
	}
	if err := g.ResolveCoinCall(flip.ID, owner.ID, "heads"); err != nil {
		t.Fatal(err)
	}
	settleMinotaurTrigger(t, g, owner.ID)

	restored := restoreRoundTrip(t, g, true)
	next := func(tbl *game.Game) (won, lost int) {
		tbl.WithWriteLock(func() {
			tbl.FlipCoinForEffect(game.CoinFlipSpec{Flipper: owner.ID, Source: minotaur, Question: "flip again"})
		})
		c := latestChoiceOfKindFor(tbl, game.PendingChoiceCoinCall, owner.ID)
		if c == nil {
			t.Fatal("no second coin call")
		}
		if err := tbl.ResolveCoinCall(c.ID, owner.ID, "tails"); err != nil {
			t.Fatal(err)
		}
		return flipEventsBy(tbl, owner.ID)
	}
	w1, l1 := next(g)
	w2, l2 := next(restored)
	if w1 != w2 || l1 != l2 {
		t.Fatalf("original %d won %d lost; restored %d won %d lost", w1, l1, w2, l2)
	}
}

// A bot sees both answers to the flip: say yes (flip) or sacrifice.
func TestKarplusanMinotaurUpkeepEnumeratorOffersBothAnswers(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	minotaurWithAge(g, owner.ID, 0)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	yes, no := false, false
	for _, m := range legal.EnumerateFor(g, owner.ID) {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		if strings.Contains(string(m.Params), "\"apply\":true") {
			yes = true
		}
		if strings.Contains(string(m.Params), "\"apply\":false") {
			no = true
		}
	}
	if !yes || !no {
		t.Errorf("enumerator offers yes=%v no=%v, want both", yes, no)
	}
}
