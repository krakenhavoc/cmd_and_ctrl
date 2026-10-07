package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exertAttackMove is attackMove's twin that also exerts the attacker
// (ADR 0130 §6).
func exertAttackMove(t *testing.T, seat int, attacker string, target int) legal.Move {
	m := attackMove(t, seat, attacker, target)
	m.Label += " and exert it (it won't untap during your next untap step)"
	m.Params = mustJSON(t, map[string]any{"attacker": attacker, "target": seatID(target).String(), "exert": true})
	return m
}

// exertRow adds a triggered ability row stamped as one of exert's
// (ADR 0130's amendment of 2026-10-07) with the given purpose.
func exertRow(kind string, p protocol.PurposeView) cardOpt {
	return func(c *protocol.CardView) {
		c.AbilityRows = append(c.AbilityRows, protocol.AbilityRowView{
			Kind: "triggered", Label: "exert " + kind, Exert: kind, Purpose: &p,
		})
	}
}

// twinInput offers pass, the exert twin (listed first, so a policy that
// read the list blindly would take it) and the plain attack.
func twinInput(t *testing.T, v protocol.GameView, attacker string) aiseat.Input {
	return input(0, v, passMove(0), exertAttackMove(t, 0, attacker, 1), attackMove(t, 0, attacker, 1))
}

func exerted(in aiseat.Input, d aiseat.Decision) bool {
	return d.Index >= 0 && strings.HasSuffix(in.Moves[d.Index].Label, "untap step)")
}

// TestBaselineNeverExerts: BaselineConfig (PriceExert off) attacks as
// the bot did before ADR 0130 §9, and passes when only the exert twin
// is on offer.
func TestBaselineNeverExerts(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(12)), newSeat(1, withLife(2))},
		withBattlefield(creature(cardID(10), 0, "Avenger", 3, 1,
			exertRow("linked", protocol.PurposeView{PreventCombatDamageToSelf: true}))),
		withTurn(9, 0, "declare_attackers"),
	)
	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	in := twinInput(t, v, cardID(10))
	if got := chose(t, in, decide(t, base, in)); got != attackMove(t, 0, cardID(10), 1).Label {
		t.Fatalf("chose %q, want the plain attack", got)
	}
	in = input(0, v, passMove(0), exertAttackMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, base, in)); got != "Pass priority" {
		t.Fatalf("chose %q; the baseline never exerts", got)
	}
}

// TestExertWithNothingToGainIsNotTaken: an exert whose rows declare
// nothing gains nothing, and keeping the creature home next turn costs
// something, so the plain attack wins.
func TestExertWithNothingToGainIsNotTaken(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(
			creature(cardID(10), 0, "Exerter", 3, 3),
			creature(cardID(20), 1, "Raider", 2, 2),
		),
		withTurn(9, 0, "declare_attackers"),
	)
	in := twinInput(t, v, cardID(10))
	d := decide(t, heuristic.New(), in)
	if exerted(in, d) {
		t.Fatalf("exerted for nothing: %s", d.Reason)
	}
}

// TestVigilanceMakesAPumpFree: a vigilant creature's exert costs
// nothing, so a linked +2/+2 that wins the combat is taken.
func TestVigilanceMakesAPumpFree(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(
			creature(cardID(10), 0, "Pumper", 2, 2, keywords("vigilance"),
				exertRow("linked", protocol.PurposeView{Pump: &protocol.PumpView{Power: 2, Toughness: 2}})),
			creature(cardID(20), 1, "Wall", 3, 3),
		),
		withTurn(9, 0, "declare_attackers"),
	)
	in := twinInput(t, v, cardID(10))
	d := decide(t, heuristic.New(), in)
	if !exerted(in, d) {
		t.Fatalf("chose %q (%s), want the exert: it is free and the pump wins the block", chose(t, in, d), d.Reason)
	}
	if !strings.Contains(d.Reason, "cost 0.00") {
		t.Errorf("reason %q: a vigilant creature's exert should cost nothing", d.Reason)
	}
}

// TestWontUntapAnywayMakesAnExertFree: a creature already marked not to
// untap during the bot's next untap step loses nothing by exerting.
func TestWontUntapAnywayMakesAnExertFree(t *testing.T) {
	marked := func(c *protocol.CardView) { c.NoUntap = &protocol.NoUntapView{Next: []string{seatID(0).String()}} }
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(
			creature(cardID(10), 0, "Pumper", 2, 2, marked,
				exertRow("linked", protocol.PurposeView{Pump: &protocol.PumpView{Power: 2, Toughness: 2}})),
			creature(cardID(20), 1, "Wall", 3, 3),
		),
		withTurn(9, 0, "declare_attackers"),
	)
	in := twinInput(t, v, cardID(10))
	if d := decide(t, heuristic.New(), in); !exerted(in, d) {
		t.Fatalf("chose %q (%s), want the free exert", chose(t, in, d), d.Reason)
	}
}

// TestGlorybringerExertsToKillTheBestCreature: four damage to the
// opponent's best creature is worth more than the dragon's next swing
// and its blocks.
func TestGlorybringerExertsToKillTheBestCreature(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(
			creature(cardID(10), 0, "Glorybringer", 4, 4, keywords("flying", "haste"),
				exertRow("linked", protocol.PurposeView{DamageToCreature: 4})),
			creature(cardID(20), 1, "Big Threat", 4, 4, keywords("trample", "lifelink")),
		),
		withTurn(9, 0, "declare_attackers"),
	)
	in := twinInput(t, v, cardID(10))
	if d := decide(t, heuristic.New(), in); !exerted(in, d) {
		t.Fatalf("chose %q (%s), want the exert that kills the 4/4", chose(t, in, d), d.Reason)
	}

	// With no creature it could kill, the trigger is worth nothing and
	// the dragon attacks plainly.
	v.Battlefield.Cards = v.Battlefield.Cards[:1]
	in = twinInput(t, v, cardID(10))
	if d := decide(t, heuristic.New(), in); exerted(in, d) {
		t.Fatalf("exerted with no target worth it: %s", d.Reason)
	}
}

// TestPayoffRowsCountOnALethalPush: a lethal push makes the exert free,
// and a "whenever you exert" payoff on ANOTHER permanent the bot
// controls is gain, so the twin is taken.
func TestPayoffRowsCountOnALethalPush(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(12)), newSeat(1, withLife(2))},
		withBattlefield(
			creature(cardID(10), 0, "Exerter", 3, 1),
			creature(cardID(11), 0, "Survivors", 3, 3, tapped(),
				exertRow("payoff", protocol.PurposeView{DamageEachOpponent: 1, LifeGain: 1})),
		),
		withTurn(9, 0, "declare_attackers"),
	)
	in := twinInput(t, v, cardID(10))
	d := decide(t, heuristic.New(), in)
	if !exerted(in, d) {
		t.Fatalf("chose %q (%s), want the free exert with a payoff", chose(t, in, d), d.Reason)
	}
}

// TestCelebrantPricesTheExtraCombat: an additional combat is worth the
// other untapped creatures' attacks in it, against an open board. The
// defender's life is out of reach of a race plan, which only ever takes
// a free exert (raceAttack), so decideAttack prices it.
func TestCelebrantPricesTheExtraCombat(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(100))},
		withBattlefield(
			creature(cardID(10), 0, "Celebrant", 4, 1,
				exertRow("linked", protocol.PurposeView{ExtraCombat: 1})),
			creature(cardID(11), 0, "Bear A", 3, 3),
			creature(cardID(12), 0, "Bear B", 3, 3),
			creature(cardID(13), 0, "Bear C", 3, 3),
		),
		withTurn(9, 0, "declare_attackers"),
	)
	in := twinInput(t, v, cardID(10))
	if d := decide(t, heuristic.New(), in); !exerted(in, d) {
		t.Fatalf("chose %q (%s), want the exert for the extra combat", chose(t, in, d), d.Reason)
	}

	// Alone, there is no second combat worth having.
	v.Battlefield.Cards = v.Battlefield.Cards[:1]
	in = twinInput(t, v, cardID(10))
	if d := decide(t, heuristic.New(), in); exerted(in, d) {
		t.Fatalf("exerted a lone Celebrant: %s", d.Reason)
	}
}
