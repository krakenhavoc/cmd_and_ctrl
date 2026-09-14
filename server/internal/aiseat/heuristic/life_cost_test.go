package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// life_cost_test.go covers #74: an activated ability's life component
// reaching the policy, and the policy doing something sane with it.
//
// The failure being fixed is not "the bot misplays a little". Before
// legal.MoveCost, every activation priced at a flat ActivateBase and a
// bot handed Griselbrand activated it until it was at 0 — it removed
// itself from the game, silently, with a legal move the engine was
// right to accept. So the two things worth pinning are opposite in
// direction: it must STOP, and it must not have become a bot that
// refuses to pay life at all, which is a different bad policy wearing
// the same fix.
//
// The whole-game version of both is in
// aiseat/life_cost_game_test.go — TestBotDoesNotPayItsLastLife and
// TestBotStillPaysLifeWhenItIsRight — which play the real engine with
// the real enumerator and the real catalog. These are the microsecond
// version that says exactly which number was wrong when one of those
// fails.

// activateMove is an activated-ability move with a declared cost,
// as legal/abilities.go emits one.
func activateMove(t *testing.T, seat int, src, label string, cost *legal.MoveCost, targets ...map[string]string) legal.Move {
	params := map[string]any{"source_card_id": src, "ability_index": 0}
	if len(targets) > 0 {
		params["targets"] = targets
	}
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src), Cost: cost,
		Params: mustJSON(t, params),
	}
}

func payLife(n int) *legal.MoveCost { return &legal.MoveCost{Life: n} }

// griselbrand is the card the bug was found on: a pure life cost, no
// tap and no mana, so it can be activated repeatedly and at any life
// total the seat can pay from — including its last.
func griselbrand(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Griselbrand", 7, 7, keywords("flying", "lifelink"))
	c.ManaCost = "{4}{B}{B}{B}{B}"
	c.TypeLine = "Legendary Creature — Demon"
	return c
}

const griselbrandAbility = "Griselbrand: Pay 7 life: Draw seven cards"

// The life total decides the play, and nothing else does: same board,
// same ability, same seven-life price.
func TestLifeCostIsPricedAgainstTheLifeTotal(t *testing.T) {
	cases := []struct {
		life int
		want bool // activate?
		why  string
	}{
		{40, true, "seven life out of forty is the cheapest card draw in Commander"},
		{20, true, "still miles from the danger zone"},
		{12, false, "paying would land on 5, deep inside DangerLife"},
		{8, false, "paying would land on 1"},
		{7, false, "paying would land on 0 — the floor, whatever the score says"},
	}
	for _, tc := range cases {
		v := newView([]protocol.PlayerView{newSeat(0, withLife(tc.life)), newSeat(1)},
			withBattlefield(griselbrand(cardID(1), 0), land(cardID(10), 0)))
		in := input(0, v,
			passMove(0),
			activateMove(t, 0, cardID(1), griselbrandAbility, payLife(7)),
		)
		got := chose(t, in, decide(t, heuristic.New(), in))
		if activated := got == griselbrandAbility; activated != tc.want {
			t.Errorf("at %d life the bot chose %q, want activate=%v — %s",
				tc.life, got, tc.want, tc.why)
		}
	}
}

// A one-life cost is the Necropotence shape, and the interesting
// thing about it is that the bot may pay it many times in a row: each
// activation is its own decision against a board one life poorer. It
// has to stop on its own, and it must not be the last point that
// stops it.
func TestRepeatedOneLifeCostStopsWellShortOfZero(t *testing.T) {
	const ability = "Necropotence: Pay 1 life: Exile the top card of your library face down."
	stopped := 0
	for life := 40; life >= 1; life-- {
		v := newView([]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)},
			withBattlefield(
				protocol.CardView{
					InstanceID: cardID(1), Name: "Necropotence", TypeLine: "Enchantment",
					Owner: seatID(0).String(), Controller: seatID(0).String(),
					ManaCost: "{B}{B}{B}", KnownByYou: true,
				},
				land(cardID(10), 0)))
		in := input(0, v, passMove(0), activateMove(t, 0, cardID(1), ability, payLife(1)))
		if chose(t, in, decide(t, heuristic.New(), in)) != ability {
			stopped = life
			break
		}
	}
	if stopped == 0 {
		t.Fatal("the bot paid one life at every total from 40 down to 1 — it would have killed itself")
	}
	if stopped < 5 {
		t.Errorf("the bot kept paying down to %d life; that is a bot trading its life for a card it cannot read", stopped)
	}
	t.Logf("stopped paying 1 life at %d life", stopped)
}

// The floor is not "the score usually works out". This is the case
// where the score emphatically does not: the ability may be pointed
// at an opponent inside FinishLife, which is worth LethalBonus — the
// single biggest number in the whole heuristic — and paying for it
// costs the bot its last seven life.
//
// Killing one of three opponents is not worth losing the game, and no
// payoff that fits on the wire is.
func TestNoPayoffBuysTheLastPointOfLife(t *testing.T) {
	v := newView([]protocol.PlayerView{
		newSeat(0, withLife(7)),
		newSeat(1, withLife(1)),
		newSeat(2),
		newSeat(3),
	}, withBattlefield(griselbrand(cardID(1), 0), land(cardID(10), 0)))
	const ability = "Demon: Pay 7 life: deal damage to any target"
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), ability, payLife(7), playerTarget(1)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == ability {
		t.Fatal("the bot paid its last seven life to point an ability at a player on 1 — it killed itself for a kill")
	}

	// And the refusal is legible rather than an accident of the
	// arithmetic: Rank carries the reason out to Layer C.
	ranked := heuristic.New().Rank(context.Background(), in)
	last := ranked[len(ranked)-1]
	if last.Index != 1 || !strings.Contains(last.Reason, "last life") {
		t.Errorf("Rank put %+v last; want the suicidal activation with a reason naming it", last)
	}
}

// LifeCostValue has to be the Strength delta the evaluation would
// compute if the life were simply gone — if it is not, a life cost is
// being weighed on a different scale from everything it competes
// against, which is exactly the class of bug this PR is fixing.
func TestLifeCostValueIsTheStrengthDelta(t *testing.T) {
	w := heuristic.DefaultWeights()
	for _, life := range []int{40, 20, 12, 10, 9, 5, 2} {
		for _, pay := range []int{1, 3, 7} {
			if pay >= life {
				continue
			}
			before := w.Evaluate(newView([]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)}))
			after := w.Evaluate(newView([]protocol.PlayerView{newSeat(0, withLife(life-pay)), newSeat(1)}))
			want := before[seatID(0).String()].Strength - after[seatID(0).String()].Strength
			got := w.LifeCostValue(life, pay)
			if diff := got - want; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("LifeCostValue(%d, %d) = %.4f, Strength delta is %.4f", life, pay, got, want)
			}
		}
	}
}

// Loyalty is the other component Params cannot name. The bot cannot
// read what an ultimate does, so the only half of the trade it can
// see is the one it pays — which is the right way round.
func TestLoyaltyCostIsPriced(t *testing.T) {
	walker := protocol.CardView{
		InstanceID: cardID(1), Name: "Planeswalker", TypeLine: "Legendary Planeswalker — Test",
		Owner: seatID(0).String(), Controller: seatID(0).String(), KnownByYou: true,
		Counters: map[string]int{"loyalty": 3},
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(walker, land(cardID(10), 0)))
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), "Planeswalker: +1", &legal.MoveCost{Loyalty: 1}),
		activateMove(t, 0, cardID(1), "Planeswalker: −3", &legal.MoveCost{Loyalty: -3}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Planeswalker: +1" {
		t.Errorf("chose %q; the +1 keeps the planeswalker and the −3 kills it", got)
	}
}
