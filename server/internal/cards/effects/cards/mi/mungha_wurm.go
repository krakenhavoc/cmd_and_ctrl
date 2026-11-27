package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mungha Wurm — Creature — Wurm {2}{G}{G}, 6/5 (Prophecy, S82 OLD b671, #5617):
//
//	"You can't untap more than one land during your untap step."
//
// Winter Orb's untap cap (ADR 0070) on lands, live only in its controller's own
// untap step: other players untap their lands as normal, and the Wurm's cap
// holds whether or not it is tapped. With two or more tapped lands the step asks
// its controller which one untaps (CR 502.3).
//
// No simplification.
func init() {
	cap := effects.CantUntapMoreThan("Mungha Wurm — no more than one land", 1, effects.Land())
	cap.Applies = func(_ *game.Game, source *game.Card, active uuid.UUID) bool {
		return source != nil && active == source.Controller
	}
	effects.Register(effects.Spec{
		OracleID:     "fe6ff20b-5e4d-44d5-b011-c6e512169294",
		Name:         "Mungha Wurm",
		Completeness: effects.CompletenessFull,
		UntapCaps:    []game.UntapCap{cap},
	})
}
