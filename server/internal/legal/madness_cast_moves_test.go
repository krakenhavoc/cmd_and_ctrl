package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// madness_cast_moves_test.go — #657: the bot half of madness.
//
// The keyword adds no enumerator code, which is the claim under test.
// The may-cast prompt already has an answer (choices.go's
// PendingChoiceMayCast arm, built for cascade), and the cast itself is
// an ordinary granted cast out of exile — so a bot that took the offer
// is offered the spell, at the madness price, with the key the engine
// validates it under.
//
// The window is the half worth a test of its own. A madness SORCERY is
// castable outside a main phase, because the grant carries CR 608.2g's
// "ignoring timing restrictions" as `TimingFlash`; a bot never offered
// it would sit on the card until the grant lapsed and the engine swept
// it into the graveyard. The same grant without the timing field is
// the control, and it is offered nothing.

func TestTheEnumeratorOffersAMadnessCastAtInstantSpeed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		timing game.GrantTiming
		want   int
	}{
		{"flash-timed, as madness grants it", game.TimingFlash, 1},
		{"without the timing, a sorcery waits", game.TimingNormal, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			seat := g.Seats[g.Turn.ActiveSeat]
			clearHand(seat)
			// newTable parks on the active seat's DRAW step: this
			// seat holds priority, and it is not a sorcery-speed
			// window (CR 307.1), which is exactly the pair the grant's
			// timing has to beat.
			battlefieldCard(g, seat, basic("Mountain", "Mountain"))

			temper := exileCardWithGrant(g, seat, game.Card{
				Name: "Test Madness Sorcery", TypeLine: "Sorcery",
				ManaCost: "{1}{R}{R}", Layout: "normal",
			}, game.CastPermission{
				Player:     seat.ID,
				Zone:       game.ZoneExile,
				AltCostKey: game.AltCostKeyMadness,
				Cost:       "{R}",
				Timing:     tc.timing,
				CastOnly:   true,
				Label:      game.MadnessTriggerLabel,
			})

			moves := castMovesFor(legal.EnumerateFor(g, seat.ID), temper)
			if len(moves) != tc.want {
				t.Fatalf("madness cast moves = %d, want %d", len(moves), tc.want)
			}
			if tc.want == 0 {
				return
			}
			var params struct {
				FromZone        string `json:"from_zone"`
				AlternativeCost string `json:"alternative_cost"`
			}
			if err := json.Unmarshal(moves[0].Params, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			if params.FromZone != "exile" {
				t.Errorf("from_zone = %q, want exile", params.FromZone)
			}
			if params.AlternativeCost != game.AltCostKeyMadness {
				t.Errorf("alternative_cost = %q, want %q", params.AlternativeCost, game.AltCostKeyMadness)
			}
		})
	}
}
