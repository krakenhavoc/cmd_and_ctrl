package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// FiredrinkerSatyrReflect is "it deals that much damage to you" for a
// creature that was dealt damage: the damaged creature (as it last existed, if
// it died to it) deals the batch's total to the trigger's controller.
func FiredrinkerSatyrReflect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return DealDamage{Source: item.Trigger.Event.Target, Target: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}

// OldB394CantAttackDuringExtraTurns is "This creature can't attack during
// extra turns" (Medomai the Ageless; CR 508.1c): a restriction on the
// creature itself, read live at declaration, that refuses every attack while
// the turn is one an effect made.
func OldB394CantAttackDuringExtraTurns() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{CantAttackDuringExtraTurns: true})
}
