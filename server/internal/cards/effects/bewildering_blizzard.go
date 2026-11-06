package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "84469042-d48d-409f-a2c9-047058f13bc5",
		Name:         "Bewildering Blizzard",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			BoostUntilEOT{
				Match: And(Creature(), OpponentControls()), Power: -3,
				Label: "Bewildering Blizzard — creatures your opponents control get -3/-0",
			},
		),
	})
}
