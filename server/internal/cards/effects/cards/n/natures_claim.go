package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Claim — Instant {G}:
//
//	"Destroy target artifact or enchantment. Its controller gains 4
//	life."
//
// One green mana for unconditional artifact/enchantment removal; the
// 4 life is the whole cost, and it goes to the PERMANENT'S
// controller, not the caster — so blowing up your own Signet gains
// you the life. Controller is read before the destroy, as on Beast
// Within.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d4e558e-9109-4918-a082-fdcbaffd516b",
		Name:         "Nature's Claim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: controller, Amount: 4}.Apply(ctx)
		},
	})
}
