package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vrondiss, Rage of Ancients — Legendary Creature — Dragon Barbarian
// {3}{R}{G}, 5/4 (Adventures in the Forgotten Realms):
//
//	"Enrage — Whenever Vrondiss is dealt damage, you may create a 5/4 red
//	 and green Dragon Spirit creature token with 'When this token deals
//	 damage, sacrifice it.'
//	 Whenever you roll one or more dice, you may have Vrondiss deal 1
//	 damage to itself."
//
// The enrage is one trigger per damaged batch (cmdr2B08EnrageTrigger); the
// token is a catalog template (printedAfrDragonSpiritToken) so its own
// trigger is on it as printed. The dice trigger is one per instruction to
// roll (WheneverYouRollDice) and does nothing once Vrondiss has left the
// battlefield.
func init() {
	effects.Register(effects.Spec{
		OracleID: "2160c837-342b-409f-a1db-0ec5ac023ccf", Name: "Vrondiss, Rage of Ancients", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Vrondiss, Rage of Ancients — Enrage: you may create a 5/4 red and green Dragon Spirit",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Vrondiss, Rage of Ancients — create a 5/4 red and green Dragon Spirit creature token?",
						Then:     vrondissDragonSpirit,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WheneverYouRollDice("Vrondiss, Rage of Ancients — you may have it deal 1 damage to itself",
				func(g *game.Game, item *game.StackItem, _ int) error {
					if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.MayChoice{
						Question: "Vrondiss, Rage of Ancients — have it deal 1 damage to itself?",
						Then:     vrondissSelfDamage,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var vrondissDragonSpirit = effects.MayYes("confirm/vrondiss-dragon-spirit", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.CreateToken{Template: effects.AfrDragonSpiritToken(), N: 1}.Apply(ctx)
})

var vrondissSelfDamage = effects.MayYes("confirm/vrondiss-self-damage", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	vrondiss := ctx.Source()
	if ctx.IsNewSourceObject(vrondiss) || !effects.OnBattlefield(ctx.Game, vrondiss) {
		return nil
	}
	return effects.DealDamage{Source: vrondiss, Target: vrondiss, Amount: 1}.Apply(ctx)
})
