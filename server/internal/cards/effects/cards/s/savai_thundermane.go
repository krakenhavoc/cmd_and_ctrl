package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savai Thundermane — Creature — Elemental Cat {R}{W}, 3/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever you cycle a card, you may pay {2}. When you do, this creature
//	 deals 2 damage to target creature and you gain 2 life."
//
// Paying {2} is made as the trigger resolves; the reflexive trigger then
// targets a creature and, as it resolves, deals the damage from Savai
// Thundermane (read from last-known information if it has left, CR 608.2h)
// and gains the life even if the creature is gone from the battlefield only
// when it was still a legal target (CR 608.2b). With no creature to target
// the reflexive trigger is removed and nothing is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "620a2037-8bf5-48e0-af94-f4f7d3650c60",
		Name:         "Savai Thundermane",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCycle(savaiThundermaneLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: savaiThundermaneLabel + "?",
					Then:     savaiThundermanePaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

const savaiThundermaneLabel = "Savai Thundermane — pay {2} to deal 2 damage to target creature and gain 2 life"

var savaiThundermanePaid = effects.PayUnlessBranches("pay-unless/savai-thundermane-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Savai Thundermane — it deals 2 damage to target creature and you gain 2 life", savaiThundermaneBody).Apply(ctx)
	}, nil)

var savaiThundermaneBody = game.ReflexiveBody("savai-thundermane/damage", effects.SimpleBody(savaiThundermaneDamage),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))

func savaiThundermaneDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Amount: 2}.Apply(ctx)
}
