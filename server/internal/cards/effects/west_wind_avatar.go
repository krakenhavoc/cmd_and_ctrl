package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// West Wind Avatar — {5}{G}{G} Creature — Cat Spirit Avatar 7/7 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 Whenever this creature enters or attacks, you may sacrifice a token or
//	 a land. If you do, you gain 3 life.
//	 Disappear — At the beginning of your end step, if a permanent left the
//	 battlefield under your control this turn, draw a card."
//
// The sacrifice is chosen as the trigger resolves, over the tokens and
// lands you control then, with a floor of zero for the "may"
// (maySacrificeOne); the life follows only a permanent that really went.
// Disappear is an ability word (CR 207.2c) for revolt's condition, an
// intervening "if" checked as the trigger fires and again as it resolves
// (AtYourEndStepIfRevolt). The sacrifice the first trigger asks for is
// itself a permanent leaving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "92e45a8f-c582-45e1-9b7f-def219023ebf",
		Name:            "West Wind Avatar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("West Wind Avatar — you may sacrifice a token or a land to gain 3 life",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "West Wind Avatar — sacrifice a token or a land to gain 3 life?",
						Or(IsTokenPredicate(), Land()), false, westWindAvatarGain)
				}),
			AtYourEndStepIfRevolt("West Wind Avatar — disappear: draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// westWindAvatarGain is "if you do, you gain 3 life", keyed (#4290).
var westWindAvatarGain = RegisterSacrificedOne("prompt-run/west-wind-avatar-gain-3", func(ctx *Context, _ uuid.UUID) error {
	return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
})
