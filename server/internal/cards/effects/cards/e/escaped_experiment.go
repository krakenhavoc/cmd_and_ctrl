package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escaped Experiment — Artifact Creature — Phyrexian Beast {1}{U}, 2/1
// (Phyrexia: All Will Be One):
//
//	"Whenever this creature attacks, target creature an opponent controls
//	 gets -X/-0 until end of turn, where X is the number of artifacts you
//	 control."
//
// X is counted as the trigger resolves (CR 608.2h), by layered type, so an
// artifact that arrived or left while it waited is counted as it is then.
// An illegal target at resolution means no effect (CR 608.2b).
//
// No simplification.
func init() {
	label := "Escaped Experiment — target creature an opponent controls gets -X/-0 until end of turn, X the number of artifacts you control"
	effects.Register(effects.Spec{
		OracleID:     "d1faaa3e-8e99-48e9-bb04-98bd46b7217c",
		Name:         "Escaped Experiment",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					x := len(effects.MatchingBattlefield(ctx, effects.And(effects.Artifact(), effects.YouControl())))
					return effects.BoostUntilEOT{Target: t.ID, Power: -x, Label: "Escaped Experiment — -X/-0"}.Apply(ctx)
				}
				return nil
			}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
