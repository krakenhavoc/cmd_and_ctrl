package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2121bcb4-f5fe-447f-9112-c7e5b5c45baf",
		Name:         "Dina, Soul Steeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Dina, Soul Steeper — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Dina gets +X/+0 until end of turn, where X is the sacrificed creature's power",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: item.SourceCardID, Power: ctx.SacrificedPower(), Label: "Dina, Soul Steeper — +X/+0 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
