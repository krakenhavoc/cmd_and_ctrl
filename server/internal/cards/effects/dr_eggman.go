package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dr. Eggman — Legendary Creature — Human Scientist {2}{U}{B}{R}, 3/6:
//
//	"Flying
//	 At the beginning of your end step, draw a card. Then each opponent
//	 faces a villainous choice — That player discards a card, or you
//	 may put a Construct, Robot, or Vehicle card from your hand onto
//	 the battlefield."
//
// A villainous choice (CR 701.55, #3140) asked of each opponent in APNAP
// order, one at a time (CR 701.55d), after the controller's draw. The
// discard is the facing player's own pick from their hand; the second
// option is the CONTROLLER's "you may" (an optional put-from-hand prompt
// addressed to them, not to the player who chose it), over the
// Construct, Robot and Vehicle permanent cards in their hand. The
// controller's hand is hidden, so the hint a bot sees on the second
// option is a fixed guess and not a read of it.
//
// No simplifications.
var drEggmanChoice = NewVillainousChoice("dr-eggman",
	"Dr. Eggman — discard a card, or the controller may put a Construct, Robot or Vehicle onto the battlefield?",
	VillainousOption{
		Label:    "Discard a card",
		Harm:     1,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return villainousHandSize(g, facing) > 0 },
		Body:     villainousDiscard(1),
	},
	VillainousOption{
		Label: "The controller of Dr. Eggman may put a Construct, Robot, or Vehicle card from their hand onto the battlefield",
		Harm:  2,
		Body: func(ctx *Context, _ Facing) error {
			return PutFromHandOntoBattlefield{
				Player:   ctx.Controller(),
				Match:    AnySubtype("Construct", "Robot", "Vehicle"),
				Optional: true,
				Label:    "Dr. Eggman — you may put a Construct, Robot, or Vehicle card from your hand onto the battlefield",
			}.Apply(ctx)
		},
	})

func init() {
	Register(Spec{
		OracleID:        "fb68c1ae-a129-4c17-81fa-be5da2c32ec5",
		Name:            "Dr. Eggman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Dr. Eggman — draw a card, then each opponent faces a villainous choice",
				// The choice waits for a prompt the draw asks (CR 608.2c, #4155).
				Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error {
						return FaceVillainousChoice{Players: ctx.Opponents(), Choice: drEggmanChoice}.Apply(ctx)
					}),
				)),
		},
	})
}
