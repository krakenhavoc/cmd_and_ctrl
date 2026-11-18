package effects

import (
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
	Register(Spec{
		OracleID: "2160c837-342b-409f-a1db-0ec5ac023ccf", Name: "Vrondiss, Rage of Ancients", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Vrondiss, Rage of Ancients — Enrage: you may create a 5/4 red and green Dragon Spirit",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Vrondiss, Rage of Ancients — create a 5/4 red and green Dragon Spirit creature token?",
						Then:     vrondissDragonSpirit,
					}.Apply(NewContext(g, item))
				}),
			WheneverYouRollDice("Vrondiss, Rage of Ancients — you may have it deal 1 damage to itself",
				func(g *game.Game, item *game.StackItem, _ int) error {
					if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return MayChoice{
						Question: "Vrondiss, Rage of Ancients — have it deal 1 damage to itself?",
						Then:     vrondissSelfDamage,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var vrondissDragonSpirit = MayYes("confirm/vrondiss-dragon-spirit", func(ctx *Context, _ game.ConfirmAnswer) error {
	return CreateToken{Template: afrDragonSpiritToken(), N: 1}.Apply(ctx)
})

var vrondissSelfDamage = MayYes("confirm/vrondiss-self-damage", func(ctx *Context, _ game.ConfirmAnswer) error {
	vrondiss := ctx.Source()
	if ctx.isNewSourceObject(vrondiss) || !onBattlefield(ctx.Game, vrondiss) {
		return nil
	}
	return DealDamage{Source: vrondiss, Target: vrondiss, Amount: 1}.Apply(ctx)
})
