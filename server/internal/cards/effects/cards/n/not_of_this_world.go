package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Not of This World — Kindred Instant — Eldrazi {7} (Rise of the Eldrazi;
// Commander Masters, S82 GAP2 b05 #5948):
//
//	"Counter target spell or ability that targets a permanent you control.
//	 This spell costs {7} less to cast if it targets a spell or ability
//	 that targets a creature you control with power 7 or greater."
//
// The target clause reads the candidate's announced targets: legal when
// one of them is a permanent on the battlefield the caster controls
// (ItemTargetsAPermanentYouControl), so a target
// that has since left fails the CR 608.2b recheck and the spell does
// nothing. The reduction reads Not of This World's own announced target
// (CR 601.2f comes after 601.2c) and, through it, that item's targets
// (ItTargetsAnItemThat): a creature the caster controls whose power, as
// it stands then, is 7 or more. A countered ability simply goes away.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5da5d8a3-009e-4d8a-977e-6ebc73ba2682",
		Name:         "Not of This World",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpellOrAbility("target spell or ability that targets a permanent you control",
			effects.ItemTargetsAPermanentYouControl(nil)),
		SelfCostModifiers: []game.CostModifier{notOfThisWorldReduction()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CounterTheFirstLegalStackTarget(ctx)
		},
	})
}

func notOfThisWorldReduction() game.CostModifier {
	m := effects.CostsLess(7, "This spell costs {7} less to cast if it targets a spell or ability that targets a creature you control with power 7 or greater.",
		effects.ItTargetsAnItemThat(effects.ItemTargetsAPermanentYouControl(func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.IsCreature() && c.CurrentPower() >= 7
		})))
	m.ReadsTargets = true
	return m
}
