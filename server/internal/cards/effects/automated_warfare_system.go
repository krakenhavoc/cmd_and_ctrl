package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Automated Warfare System — Artifact Creature — Construct {2}{B}, 2/2
// (Star Trek):
//
//	Flying
//	Whenever this creature attacks, you may sacrifice another creature or
//	artifact. If you do, draw a card.
//
// The sacrifice is chosen as the trigger resolves (maySacrificeOne), over
// the other creatures and artifacts its controller controls then, and the
// draw follows only a permanent that really went.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b28b4fe9-1a35-48c1-8818-f6f786b04e3e",
		Name:            "Automated Warfare System",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Automated Warfare System — you may sacrifice another creature or artifact to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Automated Warfare System — sacrifice another creature or artifact to draw a card?",
						Or(Creature(), Artifact()), true, automatedWarfareSystemDraw)
				}),
		},
	})
}

// automatedWarfareSystemDraw is "if you do, draw a card". An on-disk key.
var automatedWarfareSystemDraw = RegisterSacrificedOne("prompt-run/automated-warfare-system-draw", func(ctx *Context, _ uuid.UUID) error {
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
