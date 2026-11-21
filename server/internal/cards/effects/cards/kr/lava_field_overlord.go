package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lava-Field Overlord — Creature — Dragon {7}{R}{R}:
//
//	"Flying
//	 Assist (Another player can pay up to {7} of this spell's cost.)
//	 When this creature enters, it deals 4 damage to target creature an
//	 opponent controls."
//
// The ETB is Magmatic Galleon's shape: a mandatory pick over creatures
// an opponent controls, 4 damage from this creature as the source. The
// damage-to-target purpose is declared on the pick (ForTargets /
// DamageToTarget). A target that left in response is skipped (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f19a8c7e-1552-4786-a403-dd51c5cb2a7d",
		Name:            "Lava-Field Overlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.WhenThisEnters("Lava-Field Overlord — 4 damage to target creature an opponent controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							for _, t := range ctx.LegalTargets() {
								if t.Kind == game.TargetCard {
									return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
								}
							}
							return nil
						}),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.DamageToTarget(0, 4))),
		},
	})
}
