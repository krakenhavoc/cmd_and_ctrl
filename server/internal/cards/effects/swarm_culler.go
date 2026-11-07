package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swarm Culler — Creature — Insect Warrior {3}{B}, 2/4 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 Whenever this creature becomes tapped, you may sacrifice another
//	 creature or artifact. If you do, draw a card."
//
// Starfighter Pilot's "becomes tapped" watch (ThisBecameTapped over a tap
// and an attack, so a vigilance attacker asks nothing) with Beetle-Headed
// Merchants' may-sacrifice: the choice is made as the trigger resolves,
// over the controller's other creatures and artifacts, and the draw is
// its registered "if you do", so declining or having nothing to
// sacrifice draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6c28bde-10d9-471b-acd0-32a1c6195805",
		Name:            "Swarm Culler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Swarm Culler — you may sacrifice another creature or artifact to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Swarm Culler — sacrifice another creature or artifact to draw a card?",
						Or(Creature(), Artifact()), true, swarmCullerDraw)
				}),
		},
	})
}

var swarmCullerDraw = RegisterSacrificedOne("prompt-run/swarm-culler-draw", func(ctx *Context, _ uuid.UUID) error {
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
