package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swords to Plowshares — "Exile target creature. Its controller
// gains life equal to its power."
//
// Uses CurrentPower (printed + counter modifiers, clamped) so a
// +1/+1'd creature pays out correctly. We capture controller +
// power BEFORE ExileTarget moves the card — post-move the card's
// zone changes and the lookup could miss.
func init() {
	Register(Spec{
		OracleID:     "b1544f21-7e98-461b-aed5-e748b0168c52",
		Name:         "Swords to Plowshares",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			targetID := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(targetID)
			if !ok {
				return nil
			}
			controller := card.Controller
			lifeGain := card.CurrentPower()
			if err := (ExileTarget{Target: targetID}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, lifeGain)
		},
	})
}
