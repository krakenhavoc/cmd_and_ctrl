package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f19a8c7e-1552-4786-a403-dd51c5cb2a7d",
		Name:            "Lava-Field Overlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(
					WhenThisEnters("Lava-Field Overlord — 4 damage to target creature an opponent controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							for _, t := range ctx.LegalTargets() {
								if t.Kind == game.TargetCard {
									return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
								}
							}
							return nil
						}),
					TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(DamageToTarget(0, 4))),
		},
	})
}
