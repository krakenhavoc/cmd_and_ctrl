package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Gateway Express — Enchantment — Case {1}{W} (Murders at
// Karlov Manor, #4469):
//
//	"When this Case enters, choose target creature you don't control.
//	 Each creature you control deals 1 damage to that creature.
//	 To solve — Three or more creatures attacked this turn. (If
//	 unsolved, solve at the beginning of your end step.)
//	 Solved — Creatures you control get +1/+0."
//
// The enters trigger is Allies at Last's shape with every creature you
// control as a source: each deals 1 damage to the target as one damage
// event, read as it resolves, so a creature that entered in response
// deals damage too and the target takes nothing if it has left (CR
// 608.2b). The solve condition counts the distinct creatures declared as
// attackers this turn by anyone (TurnTally.Attacks), at the end step and
// again on resolution (CR 603.4). The solved line is a +1/+0 anthem that
// exists only while the Case is solved (ActiveWhen: Solved()).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42a9aeac-5b9c-4675-8c96-7c8ce7e9098c",
		Name:         "Case of the Gateway Express",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Case of the Gateway Express — each creature you control deals 1 damage to target creature you don't control",
				caseOfTheGatewayExpressDamage),
				TargetCreature("target creature you don't control", OpponentControls())),
			ToSolve("Case of the Gateway Express — to solve: three or more creatures attacked this turn",
				threeOrMoreCreaturesAttackedThisTurn),
		},
		Static: []game.StaticAbility{{
			Label:      "Creatures you control get +1/+0.",
			Layer:      game.Layer7PT,
			SubLayer:   game.SubLayer7C_Modify,
			AppliesTo:  CreaturesYouControl,
			ActiveWhen: Solved(),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
	})
}

// caseOfTheGatewayExpressDamage is the enters trigger: each creature
// the controller controls deals 1 damage to the target, all at once.
func caseOfTheGatewayExpressDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim, ok := ctx.ClauseTarget(0)
	if !ok || victim.Kind != game.TargetCard {
		return nil
	}
	var sources []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			sources = append(sources, c.InstanceID)
		}
	}
	return g.DamageInstanceForEffect(func() error {
		for _, id := range sources {
			if err := (DealDamage{Source: id, Target: victim.ID, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// threeOrMoreCreaturesAttackedThisTurn is the Case's solve condition.
func threeOrMoreCreaturesAttackedThisTurn(g *game.Game, _, _ uuid.UUID) bool {
	return g.CreaturesAttackedThisTurn() >= 3
}
