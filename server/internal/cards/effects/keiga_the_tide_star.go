package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keiga, the Tide Star — Legendary Creature — Dragon Spirit {5}{U}, 5/5
// (Tarkir: Dragonstorm Commander):
//
//	"Flying
//	 When Keiga dies, gain control of target creature."
//
// The control change has no stated duration, so it lasts indefinitely and
// does not wear off in the cleanup step (CR 611.2a, the printed ruling).
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b); Keiga herself is in the graveyard by then and
// cannot be it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "521b1598-58c6-47a0-a7b3-021d70bf2b05",
		Name:            "Keiga, the Tide Star",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Keiga, the Tide Star — gain control of target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return GainControl{
							Target:     t.ID,
							Controller: item.Controller,
							Duration:   game.IndefiniteDuration(),
							Label:      "Keiga, the Tide Star — gain control (no stated duration)",
						}.Apply(ctx)
					}
					return nil
				}),
				TargetCreature("target creature")),
		},
	})
}
