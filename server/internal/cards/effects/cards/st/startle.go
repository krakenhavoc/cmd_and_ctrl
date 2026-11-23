package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Startle — Instant {1}{U} (Midnight Hunt):
//
//	"Target creature gets -2/-0 until end of turn. Create a 2/2 black
//	 Zombie creature token with decayed. Draw a card."
//
// Three steps in printed order (CR 608.2c). A target that is gone only
// loses the first; the token and the draw still happen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7996222e-68c8-4a5c-9531-c60cdd88ae05",
		Name:         "Startle",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.WithRiders(effects.ForTargets(effects.HarmUntilEOT(0, -2, 0)), game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1}),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: 0, Label: "Startle — -2/-0 until end of turn"}.Apply(ctx)
			}),
			effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1},
			effects.DrawCards{N: 1},
		),
	})
}
