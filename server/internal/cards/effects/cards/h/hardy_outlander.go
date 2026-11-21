package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hardy Outlander — Legendary Enchantment — Background {1}{G} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "Whenever this creature attacks a player,
//	 if no opponent has more life than that player, another target creature
//	 you control gets +X/+X until end of turn, where X is this creature's
//	 power.""
//
// Guild Artisan's granted trigger (WheneverThisAttacksAPlayerNoOpponentRicher:
// an attack on a player, an intervening "if" read as the attack is declared
// and again as the trigger resolves, CR 603.4), with a target: another
// creature its controller controls, chosen as the trigger goes on the stack.
// X is the attacker's power as the ability resolves, or as it last was if it
// has left (CR 608.2h). Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const hardyOutlanderGrant = "hardy-outlander/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0a33e9b-dc1e-4643-b179-851f3c71a99d",
		Name:         "Hardy Outlander",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: hardyOutlanderGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.WheneverThisAttacksAPlayerNoOpponentRicher("Hardy Outlander — another target creature you control gets +X/+X until end of turn, where X is this creature's power",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							id, ok := effects.B16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							src, ok := g.LookupCardForEffect(item.SourceCardID)
							if !ok {
								return nil
							}
							x := src.CurrentPower()
							if x == 0 {
								return nil
							}
							return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Hardy Outlander — +X/+X"}.Apply(ctx)
						}),
					effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
			},
			Text: "Whenever this creature attacks a player, if no opponent has more life than that player, another target creature you control gets +X/+X until end of turn, where X is this creature's power.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(hardyOutlanderGrant)},
	})
}
