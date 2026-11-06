package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benthic Criminologists — Creature — Merfolk Wizard {4}{U}, 4/5
// (Murders at Karlov Manor, #4469):
//
//	"Whenever this creature enters or attacks, you may sacrifice an
//	 artifact. If you do, draw a card."
//
// One ability with two conditions (WhenThisEntersOrAttacks). The
// sacrifice is asked as the trigger resolves, over the artifacts its
// controller controls then (maySacrificeOne, a floor of zero for the
// "may"), and the draw is the sacrifice's continuation, so it happens
// only once an artifact has really gone ("if you do"). With no artifact
// there is no question and no draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5539d78c-1acc-46dd-ba6d-cc0cdad6ff44",
		Name:         "Benthic Criminologists",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Benthic Criminologists — you may sacrifice an artifact to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Benthic Criminologists — sacrifice an artifact to draw a card?",
						Artifact(), false, benthicCriminologistsDraw)
				}),
		},
	})
}

// benthicCriminologistsDraw is "if you do, draw a card", keyed.
var benthicCriminologistsDraw = RegisterSacrificedOne("prompt-run/benthic-criminologists-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return DrawCards{N: 1}.Apply(ctx)
})
