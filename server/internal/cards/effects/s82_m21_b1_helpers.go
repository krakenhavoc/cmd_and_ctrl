package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_m21_b1_helpers.go — shared pieces of Core Set 2021 batch 1 (#5511).

// s82M21B1ExileThenItsControllerCreates is "Exile target <permanent>. Its
// controller creates <token>." as a spell's whole body (Angelic Ascension,
// Secure the Scene): the controller is read BEFORE the exile moves the card
// (CR 608.2h; an exiled card carries no controller), and a target that left
// or became illegal in response does nothing at all (CR 608.2b).
func s82M21B1ExileThenItsControllerCreates(token string) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		target := item.Targets[0].ID
		controller, ok := controllerOfTarget(ctx, target)
		if !ok {
			return nil
		}
		if err := (ExileTarget{Target: target}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Controller: controller, Template: TokenCard(token), N: 1}.Apply(ctx)
	}
}
