package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Painsmith — Creature — Human Artificer {1}{B}, 2/1 (Scars of Mirrodin, S82
// OLD batch 472, #5577):
//
//	"Whenever you cast an artifact spell, you may have target creature get +2/+0
//	 and gain deathtouch until end of turn."
//
// Batterhorn's optional targeted trigger on Myrsmith's artifact-cast watch: the
// creature is chosen as the trigger goes on the stack and the "may" is asked as
// it resolves. The pump and the keyword are two until-end-of-turn effects on the
// target, each skipped if it is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbdbea75-7f80-4704-b2cf-920d00dc1273",
		Name:         "Painsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverYouCast(effects.Artifact(),
				"Painsmith — you may have target creature get +2/+0 and gain deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Painsmith — +2/+0 until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Painsmith — deathtouch until end of turn"}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
				"Painsmith — give the target creature +2/+0 and deathtouch?"),
		},
	})
}
