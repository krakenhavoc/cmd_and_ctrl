package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mysterious Tome // Chilling Chronicle — Artifact — Book {2}{U} (Midnight
// Hunt, Innistrad: Double Feature):
//
//	Mysterious Tome
//	  "{2}, {T}: Draw a card. Transform this artifact."
//	Chilling Chronicle
//	  "{1}, {T}: Tap target nonland permanent. Transform this artifact."
//
// Each face's activation does its first sentence, then turns the permanent
// over in place. The transform is part of the resolving ability, so the
// permanent must still be the same object (ADR 0152 §3): a Tome that left
// and returned is not turned over. The tap target is chosen on activation
// and the ability does nothing if it is gone (CR 608.2b), though the
// artifact still transforms.
//
// No simplification.
func init() {
	const oracle = "b32253d6-5f00-40a8-a7e5-dc4655cd288c"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Mysterious Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card. Transform this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.MidB12DrawThenTransformThis,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Chilling Chronicle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target nonland permanent. Transform this artifact.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.LegalTargetIDs(ctx) {
					if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.TransformThis{}.Apply(ctx)
			},
		}},
	})
}
