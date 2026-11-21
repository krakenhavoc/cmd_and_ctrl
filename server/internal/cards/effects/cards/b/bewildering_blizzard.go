package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bewildering Blizzard — Instant {4}{U}{U} (Tarkir: Dragonstorm, #4463):
//
//	"Draw three cards. Creatures your opponents control get -3/-0 until end
//	 of turn."
//
// The draw, then a power-only shrink over the creatures your opponents
// control as the spell resolves (CR 611.2c): a creature that enters
// afterwards keeps its power. Toughness is untouched, so nothing dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84469042-d48d-409f-a2c9-047058f13bc5",
		Name:         "Bewildering Blizzard",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -3,
				Label: "Bewildering Blizzard — creatures your opponents control get -3/-0",
			},
		),
	})
}
