package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle-Rattle Shaman — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "72dc807b-35fb-4d69-ae4c-04eed0646170", Name: "Battle-Rattle Shaman", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Optional(effects.TriggerWithPurpose(effects.Targeting(effects.AtBeginningOfYourCombat("Battle-Rattle Shaman — target creature gets +2/+0", func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Battle-Rattle Shaman — +2/+0"}).Apply(ctx)
		}), effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0))), "Battle-Rattle Shaman — give the creature +2/+0?")},
	})
}
