package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// target_purpose_test.go is ADR 0126's amendment of 2026-10-08, PR 3
// (A1, B1): a target priced by what the declared purpose does to it.

func targetEntries(ts ...protocol.TargetPurposeView) *protocol.PurposeView {
	return &protocol.PurposeView{Targets: &ts}
}

// prismari is Prismari Command's modes as PR 2 declares them.
func prismari(id string) protocol.CardView {
	c := spell(id, 0, "Prismari Command", "{1}{U}{R}")
	c.Modes = &protocol.ModeSpecView{Min: 2, Max: 2, Options: []protocol.ModeOptionView{
		{Label: "Prismari Command deals 2 damage to any target.", Purpose: targetEntries(protocol.TargetPurposeView{Slot: 0, Damage: 2})},
		{Label: "Target player draws two cards, then discards two cards.", Purpose: targetEntries(protocol.TargetPurposeView{Slot: 0, Draws: 2, Discards: 2})},
		{Label: "Target player creates a Treasure token.", Purpose: targetEntries(protocol.TargetPurposeView{Slot: 0, Tokens: 1})},
		{Label: "Destroy target artifact."},
	}}
	return c
}

// modalCast is a cast of `id` with these modes, one player target per
// mode occurrence, as the enumerator shapes it (targets[].mode).
func modalCast(t *testing.T, id, label string, modes []int, seats ...int) legal.Move {
	t.Helper()
	m := castMove(t, 0, id, label)
	targets := make([]map[string]any, len(seats))
	for i, s := range seats {
		targets[i] = map[string]any{"kind": "player", "id": seatID(s).String(), "mode": i}
	}
	m.Params = mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "modes": modes, "targets": targets})
	return m
}

func withoutTargetPurposes() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceTargetPurposes = false
	return heuristic.NewWithConfig(cfg)
}

// Prismari's loot and Treasure are worth more on the bot than on an
// opponent, and the amendment's worked number for "both at the bot".
func TestATargetedGiftIsWorthMoreOnTheBot(t *testing.T) {
	id := cardID(1)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(prismari(id)), withLife(34)), newSeat(1, withLife(41))},
		withBattlefield(manaLands(3, 0, 100)...), withTurn(6, 0, "precombat_main"))
	mine := modalCast(t, id, "loot me, Treasure me", []int{1, 2}, 0, 0)
	lootMe := modalCast(t, id, "loot me, Treasure them", []int{1, 2}, 0, 1)
	lootThem := modalCast(t, id, "loot them, Treasure me", []int{1, 2}, 1, 0)
	theirs := modalCast(t, id, "loot them, Treasure them", []int{1, 2}, 1, 1)
	in := input(0, v, passMove(0), mine, lootMe, lootThem, theirs)

	pol := heuristic.New()
	cfg := pol.Config()
	loot := 2*cfg.Weights.Hand - 2*cfg.DiscardWeight
	treasure := cfg.TokenWeight
	// Two seats: a gift to the only opponent costs OpponentMean +
	// OpponentMax of it.
	opp := cfg.Weights.OpponentMean + cfg.Weights.OpponentMax
	want := map[string]float64{
		mine.Label:     loot + treasure - cfg.Weights.Hand,
		lootMe.Label:   loot - opp*treasure - cfg.Weights.Hand,
		lootThem.Label: -opp*loot + treasure - cfg.Weights.Hand,
		theirs.Label:   -opp*(loot+treasure) - cfg.Weights.Hand,
	}
	for label, w := range want {
		if got := rankValue(t, pol, in, label); !nearly(got, w) {
			t.Errorf("%s priced %.3f, want %.3f", label, got, w)
		}
	}
	if got := rankValue(t, pol, in, mine.Label); !nearly(got, 0.50) {
		t.Errorf("loot and Treasure at the bot priced %.3f, want the amendment's 0.50", got)
	}
	if !(rankValue(t, pol, in, mine.Label) > rankValue(t, pol, in, lootMe.Label) &&
		rankValue(t, pol, in, lootMe.Label) > rankValue(t, pol, in, theirs.Label) &&
		rankValue(t, pol, in, lootThem.Label) > rankValue(t, pol, in, theirs.Label)) {
		t.Error("a gift at the bot does not beat the same gift at the opponent")
	}

	// With the knob off the bot is the worst target of all.
	off := withoutTargetPurposes()
	if rankValue(t, off, in, mine.Label) > rankValue(t, off, in, theirs.Label) {
		t.Error("with PriceTargetPurposes off the gift at the bot should still price below the gift at the opponent")
	}
}

// Sign in Blood at itself beats Sign in Blood at an opponent on 30.
func TestSignInBloodAtItself(t *testing.T) {
	id := cardID(1)
	sib := sorcery(id, 0, "Sign in Blood", "{B}{B}", targetEntries(protocol.TargetPurposeView{Slot: 0, Draws: 2, LifeLoss: 2}))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(sib), withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(manaLands(2, 0, 100)...), withTurn(5, 0, "precombat_main"))
	me := castMove(t, 0, id, "Sign in Blood at me", playerTarget(0))
	them := castMove(t, 0, id, "Sign in Blood at them", playerTarget(1))
	in := input(0, v, passMove(0), me, them)

	pol := heuristic.New()
	cfg := pol.Config()
	self := 2*cfg.Weights.Hand - 2*cfg.Weights.Life - cfg.Weights.Hand
	if got := rankValue(t, pol, in, me.Label); !nearly(got, self) {
		t.Errorf("Sign in Blood at the bot priced %.3f, want %.3f", got, self)
	}
	if m, o := rankValue(t, pol, in, me.Label), rankValue(t, pol, in, them.Label); m <= o {
		t.Errorf("Sign in Blood at the bot %.3f is no better than at the opponent %.3f", m, o)
	}
	off := withoutTargetPurposes()
	if m, o := rankValue(t, off, in, me.Label), rankValue(t, off, in, them.Label); m > o {
		t.Errorf("with PriceTargetPurposes off, at the bot %.3f should price below at the opponent %.3f", m, o)
	}
}

// Owner answer 2: at a four-seat table a gift to the strongest
// opponent costs OpponentMean/3 + OpponentMax of it, and one to an
// opponent who stays below the strongest only OpponentMean/3.
func TestAGiftIsWeighedByTheTable(t *testing.T) {
	id := cardID(1)
	draw := sorcery(id, 0, "Words", "{1}{U}", targetEntries(protocol.TargetPurposeView{Slot: 0, Draws: 2}))
	bf := append(manaLands(2, 0, 100), bigArmy(1, 3, 200)...)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(draw)), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(5, 0, "precombat_main"))
	leader := castMove(t, 0, id, "Words at the leader", playerTarget(1))
	other := castMove(t, 0, id, "Words at another", playerTarget(2))
	in := input(0, v, passMove(0), leader, other)

	pol := heuristic.New()
	w := pol.Config().Weights
	x := 2 * w.Hand
	if got, want := rankValue(t, pol, in, leader.Label), -(w.OpponentMean/3+w.OpponentMax)*x-w.Hand; !nearly(got, want) {
		t.Errorf("a gift to the leader priced %.3f, want %.3f", got, want)
	}
	if got, want := rankValue(t, pol, in, other.Label), -(w.OpponentMean/3)*x-w.Hand; !nearly(got, want) {
		t.Errorf("a gift to another opponent priced %.3f, want %.3f", got, want)
	}
}

// A pick with no entry, or with only a damage entry, keeps
// targetsValue's price: damage is PR 4's (DamageByLethality).
func TestDamageEntriesKeepTodaysPrice(t *testing.T) {
	id := cardID(1)
	shock := spell(id, 0, "Shock", "{R}")
	shock.Purpose = targetEntries(protocol.TargetPurposeView{Slot: 0, Damage: 2})
	bear := creature(cardID(200), 1, "Bear", 2, 4)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(shock)), newSeat(1)},
		withBattlefield(append(manaLands(1, 0, 100), bear)...), withTurn(5, 0, "precombat_main"))
	moves := []legal.Move{passMove(0),
		castMove(t, 0, id, "Shock the bear", cardTarget(bear.InstanceID)),
		castMove(t, 0, id, "Shock them", playerTarget(1)),
		castMove(t, 0, id, "Shock me", playerTarget(0)),
	}
	in := input(0, v, moves...)
	on, off := heuristic.New(), withoutTargetPurposes()
	for _, m := range moves[1:] {
		if a, b := rankValue(t, on, in, m.Label), rankValue(t, off, in, m.Label); !nearly(a, b) {
			t.Errorf("%s priced %.3f with target purposes, %.3f without", m.Label, a, b)
		}
	}

	// Prismari's damage mode beside its loot keeps the attack price at
	// the opponent; the loot at the bot is the loot.
	pid := cardID(2)
	v = newView([]protocol.PlayerView{newSeat(0, withHand(prismari(pid))), newSeat(1)},
		withBattlefield(manaLands(3, 0, 100)...), withTurn(5, 0, "precombat_main"))
	m := modalCast(t, pid, "2 at them, loot me", []int{0, 1}, 1, 0)
	cfg := on.Config()
	want := cfg.DamageToPlayer*cfg.LeaderBoost + 2*cfg.Weights.Hand - 2*cfg.DiscardWeight - cfg.Weights.Hand
	if got := rankValue(t, on, input(0, v, passMove(0), m), m.Label); !nearly(got, want) {
		t.Errorf("2 at the opponent and loot the bot priced %.3f, want %.3f", got, want)
	}
}
