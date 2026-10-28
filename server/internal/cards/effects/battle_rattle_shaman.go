package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle-Rattle Shaman — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "72dc807b-35fb-4d69-ae4c-04eed0646170", Name: "Battle-Rattle Shaman", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Optional(TriggerWithPurpose(Targeting(AtBeginningOfYourCombat("Battle-Rattle Shaman — target creature gets +2/+0", func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return (BoostUntilEOT{Target: id, Power: 2, Label: "Battle-Rattle Shaman — +2/+0"}).Apply(ctx)
		}), TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 2, 0))), "Battle-Rattle Shaman — give the creature +2/+0?")},
	})
}
