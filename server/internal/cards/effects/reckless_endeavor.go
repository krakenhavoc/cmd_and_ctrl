package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Endeavor — "Roll two d12 and choose one result. Reckless
// Endeavor deals damage equal to that result to each creature. Then create
// a number of Treasure tokens equal to the other result."
func init() {
	Register(Spec{
		OracleID: "928bf0d9-89be-462d-aeb7-74a65e95535c", Name: "Reckless Endeavor", Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			rolls, err := rollDice(ctx, 12, 2)
			if err != nil {
				return err
			}
			controller, source := item.Controller, item.SourceCardID
			apply := func(damage, treasures int) func(*game.Game) error {
				return func(g *game.Game) error {
					var targets []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() {
							targets = append(targets, c.InstanceID)
						}
					}
					return g.DealDamageEachThenForEffect(source, targets, damage, func(g *game.Game, _ int) error { return treasureTokens(g, controller, treasures) })
				}
			}
			if rolls[0] == rolls[1] {
				return apply(rolls[0], rolls[1])(ctx.Game)
			}
			ctx.Game.QueueConfirmForEffect(game.ConfirmPrompt{
				Chooser: controller, Source: source, Question: "Reckless Endeavor — choose the damage result",
				AcceptLabel:  fmt.Sprintf("%d damage, %d Treasures", rolls[0], rolls[1]),
				DeclineLabel: fmt.Sprintf("%d damage, %d Treasures", rolls[1], rolls[0]),
				OnAccept:     apply(rolls[0], rolls[1]), OnDecline: apply(rolls[1], rolls[0]),
			})
			return nil
		},
	})
}
