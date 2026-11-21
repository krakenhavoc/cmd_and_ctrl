package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Defiler — Artifact Creature — Demon Construct {3}{B}{R}, 5/4:
//
//	"Trample
//	 Battle Cannon — When this creature enters or dies, for each opponent,
//	 choose a nonland permanent that player controls. Destroy one of them
//	 chosen at random."
//
// You (the controller) make one pick per opponent still in the game; nothing
// targets, so hexproof and protection don't narrow the choice. Then one of
// the picks that is still on the battlefield is destroyed, at random. A
// controller whose opponents have no nonland permanent picks nothing.
var chaosDefilerPick = effects.RegisterPermanentsPick("prompt-run/chaos-defiler-destroy", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		var live []uuid.UUID
		for _, id := range picked.Cards() {
			if effects.OnBattlefield(ctx.Game, id) {
				live = append(live, id)
			}
		}
		chosen := ctx.Game.ChooseAtRandomForEffect(game.RandomDraw{Player: ctx.Controller(), Source: ctx.Source()}, live, 1)
		if len(chosen) == 0 {
			return nil
		}
		return effects.DestroyTarget{Target: chosen[0]}.Apply(ctx)
	},
})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "eba7ef08-f444-4d23-92b8-29d2ac08b156",
		Name:            "Chaos Defiler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEntersOrDies("Chaos Defiler — for each opponent choose a nonland permanent; destroy one of them at random",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ChoosePermanents{
						Of:       ctx.Opponents(),
						Question: "Chaos Defiler — choose a nonland permanent this opponent controls",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							ids := effects.NonlandPermanentsControlledBy(g, of)
							return ids, 1, 1
						},
						Keyed: chaosDefilerPick,
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerRemove}),
		},
	})
}
