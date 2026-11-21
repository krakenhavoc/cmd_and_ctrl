package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jet, Freedom Fighter — Legendary Creature — Human Rebel Ally
// {2}{R/W}{R/W}{R/W}, 3/1 (Avatar: The Last Airbender, #4459):
//
//	"When Jet enters, he deals damage equal to the number of creatures you
//	 control to target creature an opponent controls.
//	 When Jet dies, put a +1/+1 counter on each of up to two target
//	 creatures."
//
// The damage is counted as the ability resolves (the ruling), Jet
// included while he is still there; Jet is the source, as he last existed
// if he has left (CR 608.2h). The dies trigger is Rishkar's counter on
// each of up to two targets, any creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77cd2246-6906-4dce-8580-4c9d2e68c477",
		Name:         "Jet, Freedom Fighter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Jet, Freedom Fighter — damage equal to the creatures you control to target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID,
								Amount: effects.B04CreaturesControlled(g, ctx.Controller())}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.Targeting(effects.WhenThisDies("Jet, Freedom Fighter — put a +1/+1 counter on each of up to two target creatures",
				effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
	})
}
