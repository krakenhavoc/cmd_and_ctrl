package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// impulse_test.go — #2480: Professional Face-Breaker's "Sacrifice a
// Treasure: Exile the top card of your library. You may play that card
// this turn." was never taken in 1,070 windows of ADR 0126's run 1.

const faceBreakerLabel = "Sacrifice a Treasure: Exile the top card of your library. You may play that card this turn."

func faceBreaker(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := creature(id, controller, "Professional Face-Breaker", 2, 3, opts...)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: faceBreakerLabel,
		Purpose: &protocol.PurposeView{Impulse: 1, Answers: &[]string{"value"}},
	}}
	return c
}

func treasure(id string, controller int) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Treasure", Owner: seatID(controller).String(), Controller: seatID(controller).String(),
		TypeLine: "Token Artifact — Treasure", IsToken: true, KnownByYou: true,
		ManaAbilities: []protocol.ManaAbilityView{{Index: 0, TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}},
	}
}

func impulseMove(t *testing.T, seat int, src, sacrificed string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: "Professional Face-Breaker: " + faceBreakerLabel, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0, "sacrifice_ids": []string{sacrificed}}),
	}
}

// impulseRank prices the activation and passing for seat 0 in `step` of
// seat `active`'s turn, with `lands` untapped lands and the land drop
// already made or not.
func impulseRank(t *testing.T, cfg heuristic.Config, active int, step string, lands int, landPlayed bool) (act, pass float64) {
	t.Helper()
	bf := []protocol.CardView{faceBreaker(cardID(1), 0), treasure(cardID(2), 0)}
	for i := range lands {
		bf = append(bf, land(cardID(10+i), 0))
	}
	played := func(p *protocol.PlayerView) {
		p.LandDropsPerTurn = 1
		if landPlayed {
			p.LandsPlayedThisTurn = 1
		}
	}
	v := newView([]protocol.PlayerView{newSeat(0, played), newSeat(1)}, withBattlefield(bf...), withTurn(5, active, step))
	in := input(0, v, passMove(0), impulseMove(t, 0, cardID(1), cardID(2)))
	vals := map[int]float64{}
	for _, c := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[c.Index] = c.Value
	}
	return vals[1], vals[0]
}

// In its own main phase, with mana to play what it finds, the bot
// cracks the Treasure: 1.20 for the card, less 0.50 for the Treasure
// and 0.30 for the untapped creature.
func TestImpulseIsTakenInTheOwnMainPhase(t *testing.T) {
	act, pass := impulseRank(t, heuristic.DefaultConfig(), 0, "precombat_main", 3, true)
	if act <= heuristic.DefaultConfig().PassThreshold || act <= pass {
		t.Errorf("impulse for a Treasure in the own main phase with 3 lands: %.2f against pass %.2f — want it taken", act, pass)
	}
	if want := 1.2 - 0.5 - 0.3; act < want-1e-9 || act > want+1e-9 {
		t.Errorf("impulse priced %.2f, want %.2f", act, want)
	}
}

// With no land drop left and no mana to play what it finds, or in
// another seat's turn, the card would be lost: not worth the Treasure.
func TestImpulseIsNotTakenWhenTheCardWouldBeLost(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	if act, _ := impulseRank(t, cfg, 0, "precombat_main", 1, true); act >= cfg.PassThreshold {
		t.Errorf("tapped out with the land drop made: impulse %.2f, want below PassThreshold", act)
	}
	if act, _ := impulseRank(t, cfg, 0, "precombat_main", 1, false); act <= cfg.PassThreshold {
		t.Errorf("land drop still to make: impulse %.2f, want above PassThreshold", act)
	}
	if act, _ := impulseRank(t, cfg, 1, "precombat_main", 3, false); act >= 0 {
		t.Errorf("in an opponent's main phase: impulse %.2f, want below zero", act)
	}
}

// Each switch alone leaves it under zero, as before #2480.
func TestImpulseNeedsBothSwitches(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  func(*heuristic.Config)
	}{
		{"PriceImpulse off", func(c *heuristic.Config) { c.PriceImpulse = false }},
		{"OneShotManaAsToken off", func(c *heuristic.Config) { c.OneShotManaAsToken = false }},
	} {
		cfg := heuristic.DefaultConfig()
		tc.set(&cfg)
		if act, _ := impulseRank(t, cfg, 0, "precombat_main", 3, true); act >= cfg.PassThreshold {
			t.Errorf("%s: impulse %.2f, want below PassThreshold", tc.name, act)
		}
	}
}

func TestBaselineHasNoImpulsePrice(t *testing.T) {
	c := heuristic.BaselineConfig()
	if c.PriceImpulse || c.OneShotManaAsToken {
		t.Errorf("BaselineConfig has PriceImpulse %v, OneShotManaAsToken %v; want both off", c.PriceImpulse, c.OneShotManaAsToken)
	}
}
