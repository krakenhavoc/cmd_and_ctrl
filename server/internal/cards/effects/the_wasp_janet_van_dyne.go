package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Wasp, Janet Van Dyne — {2}{W} Legendary Creature — Human Hero 1/1
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 When The Wasp enters, she deals 4 damage to target tapped creature an
//	 opponent controls."
//
// The target is chosen as the trigger is put on the stack and re-checked as
// it resolves (CR 608.2b). The damage is dealt by The Wasp even if she has
// left the battlefield by then, using her last-known information
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d35c7c8-1342-40b7-82ec-2f7973ee8af8",
		Name:            "The Wasp, Janet Van Dyne",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(
					WhenThisEnters("The Wasp — she deals 4 damage to target tapped creature an opponent controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							ts := ctx.LegalTargets()
							if len(ts) == 0 {
								return nil
							}
							return DealDamage{Source: ctx.Source(), Target: ts[0].ID, Amount: 4}.Apply(ctx)
						}),
					TargetCreature("target tapped creature an opponent controls", OpponentControls(), b751Tapped())),
				ForTargets(DamageToTarget(0, 4))),
		},
	})
}
