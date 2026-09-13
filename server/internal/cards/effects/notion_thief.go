package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Notion Thief — Creature — Human Rogue {2}{U}{B}, 3/1 (EDHREC rank
// 1190):
//
//	"Flash
//	 If an opponent would draw a card except the first one they draw
//	 in each of their draw steps, instead that player skips that draw
//	 and you draw a card."
//
// The wheel deck's mirror. A draw REPLACEMENT, not a trigger: the
// pipeline hands every draw to the Thief before it happens, and the
// Thief rewrites the drawing player to its controller — which is
// exactly "that player skips that draw and you draw a card" in one
// move, with the draw event then attributed to the Thief's
// controller, so a "whenever you draw" payoff sees the right player.
// Flash rides PrintedKeywords. Two Thieves under different
// controllers chain through the CR 616 loop the way the printed
// cards do, and CR 614.5's once-per-event tracking stops the chain
// from looping.
//
// Sandbox simplification, declared: "except the FIRST one they draw
// in each of their draw steps" is read as "except any draw during
// their own draw step". The engine keeps no per-step draw tally, so
// a second draw in the draw step — an instant cast there, a
// draw-step trigger — is left alone rather than stolen. Weaker than
// printed for the Thief's controller, never stronger.
func init() {
	Register(Spec{
		OracleID:        "f8dab16e-1d50-443e-9431-8b6f1cf61c9c",
		Name:            "Notion Thief",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Every card an opponent draws during their own draw step is left alone, not just the first one."},
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDraw || ev.DrawPlayer == src.Controller {
					return false
				}
				if g.PlayerByIDForEffect(ev.DrawPlayer) == nil {
					return false
				}
				if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
					g.Seats[g.Turn.ActiveSeat] != nil && g.Seats[g.Turn.ActiveSeat].ID == ev.DrawPlayer {
					return false
				}
				return true
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				ev.DrawPlayer = src.Controller
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Notion Thief: you draw instead",
		}},
	})
}
