package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yore-Tiller Nephilim — Creature — Nephilim {W}{U}{B}{R}, 2/2 (Guildpact,
// S82 OLD batch 573, #5599):
//
//	"Whenever this creature attacks, return target creature card from your
//	 graveyard to the battlefield tapped and attacking."
//
// Dauntless Avenger's attack trigger without its mana value limit: the
// card is put onto the battlefield attacking the player or planeswalker the
// Nephilim attacks (the controller chooses, CR 508.4), and was never declared
// as an attacker.
//
// No simplification.
func init() {
	trigger := effects.WheneverThisAttacks("Yore-Tiller Nephilim — return target creature card from your graveyard to the battlefield tapped and attacking",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, id := range effects.LegalTargetIDs(ctx) {
				return effects.PutAttackingYourChoice(ctx, id, game.ZoneGraveyard, "Yore-Tiller Nephilim")
			}
			return nil
		})
	trigger.Targets = effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature())
	effects.Register(effects.Spec{
		OracleID:     "7e9d61d3-6ba7-45ac-a55c-a2f36595c5d2",
		Name:         "Yore-Tiller Nephilim",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
