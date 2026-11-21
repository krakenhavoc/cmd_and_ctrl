package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dina, Soul Steeper — Legendary Creature — Dryad Druid {B}{G}, 1/3
// (Secrets of Strixhaven Commander, #4834):
//
//	"Whenever you gain life, each opponent loses 1 life.
//	 {1}, Sacrifice another creature: Dina gets +X/+0 until end of turn,
//	 where X is the sacrificed creature's power."
//
// The first ability is Marauding Blight-Priest's. X is the sacrificed
// creature's power as it last existed on the battlefield (CR 608.2h,
// Context.SacrificedPower), floored at zero; the pump lands only on the
// Dina the ability came from, not on one that left and came back (CR
// 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2121bcb4-f5fe-447f-9112-c7e5b5c45baf",
		Name:         "Dina, Soul Steeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Dina, Soul Steeper — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Dina gets +X/+0 until end of turn, where X is the sacrificed creature's power",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: ctx.SacrificedPower(), Label: "Dina, Soul Steeper — +X/+0 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
