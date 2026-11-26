package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ironclad Revolutionary — {4}{B}{B} Creature — Aetherborn Artificer 4/4
// (Aether Revolt, #5536):
//
//	"When this creature enters, you may sacrifice an artifact. If you do,
//	 put two +1/+1 counters on this creature and each opponent loses 2 life."
//
// "If you do" is the sacrifice's keyed continuation: it runs only once an
// artifact has really left, and the counters go on this creature only if it
// is still the permanent that entered (CR 400.7).
//
// No simplification.
var ironcladRevolutionarySacrificed effects.SacrificedOne

func init() {
	ironcladRevolutionarySacrificed = effects.RegisterSacrificedOne("prompt-run/ironclad-revolutionary-sacrifice",
		func(ctx *effects.Context, _ uuid.UUID) error {
			if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.EachOpponentLosesLife{N: 2}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "382109c5-fd08-4e82-b898-2a387181b470",
		Name:         "Ironclad Revolutionary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ironclad Revolutionary — you may sacrifice an artifact; if you do, two +1/+1 counters and each opponent loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Ironclad Revolutionary — sacrifice an artifact?",
						effects.Artifact(), false, ironcladRevolutionarySacrificed)
				}),
		},
	})
}
