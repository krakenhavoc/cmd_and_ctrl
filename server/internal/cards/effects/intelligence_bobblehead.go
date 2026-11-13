package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intelligence Bobblehead — Artifact — Bobblehead {3} (Fallout, #5469):
//
//	"{T}: Add one mana of any color.
//	 {5}, {T}: Draw X cards, where X is the number of Bobbleheads you
//	 control."
//
// Coalition Relic's any-color mana ability, and a draw whose X is
// counted as the ability resolves (CR 608.2h): a Bobblehead that leaves
// in response is not counted, and the tapped Bobblehead itself is if it
// is still there. Bobblehead is an artifact type, read off the
// permanents' effective type lines (pipBobbleheadsYouControl).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16e6bd4b-fbf1-4bbe-9271-a74f59320e0d",
		Name:         "Intelligence Bobblehead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Draw X cards, where X is the number of Bobbleheads you control.",
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: Do(Instruction(func(ctx *Context) error {
				return DrawCards{N: pipBobbleheadsYouControl(ctx.Game, ctx.Controller())}.Apply(ctx)
			})),
		}},
	})
}

// pipBobbleheadsYouControl is the Fallout Bobbleheads' "the number of
// Bobbleheads you control": every permanent `player` controls with the
// Bobblehead artifact subtype, the counting Bobblehead included.
func pipBobbleheadsYouControl(g *game.Game, player uuid.UUID) int {
	return countControlled(g, player, func(c game.Card) bool { return c.HasSubtype("Bobblehead") })
}
