package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Redirect Lightning — Instant — Lesson {R}:
//
//	"As an additional cost to cast this spell, pay 5 life or pay {2}.
//	 Change the target of target spell or ability with a single target."
//
// The either/or cost (ADR 0100 §2): the life branch is refused at
// announce below 5 life (CR 119.4), and the {2} joins the total at
// CR 601.2f. The retarget is Bolt Bend's (#1211): the new target must be
// legal for the spell or ability, and the controller of Redirect
// Lightning chooses it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c62b5c22-e058-436f-9575-a01cb2112829",
		Name:         "Redirect Lightning",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.PayLifeCost(5).Keyed("life"),
			effects.ManaAdditionalCost("{2}").Keyed("mana"),
		),
		Targets: effects.TargetSpellOrAbility("target spell or ability with a single target",
			effects.ItemHasASingleTarget()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.ChangeTargets{
				StackID: item.Targets[0].ID,
				Policy:  game.RetargetChangeOne,
				Reason:  "Redirect Lightning — change the target",
			}.Apply(ctx)
		},
	})
}
