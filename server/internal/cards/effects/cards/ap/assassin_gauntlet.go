package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assassin Gauntlet — Artifact — Equipment {2}{U} (Assassin's Creed, #5948):
//
//	"When this Equipment enters, attach it to up to one target creature you
//	 control. Tap all creatures target opponent controls.
//	 Equipped creature gets +1/+1 and has "Whenever this creature deals combat
//	 damage to a player, draw a card, then discard a card."
//	 Equip {2}"
//
// One entry trigger with two target clauses (#764): slot 0 an "up to one"
// creature you control, which the Gauntlet attaches to, and slot 1 an
// opponent, whose creatures are all tapped as it resolves (a creature that
// enters later is not). The attach goes to slot 0 only. The loot is Biorganic
// Carapace's granted combat-damage trigger over a draw-then-discard, so its
// source is the equipped creature and a removal of the Gauntlet takes it.
//
// No simplification.
func init() {
	const grantKey = "assassin-gauntlet/combat-loot"
	effects.Register(effects.Spec{
		OracleID:     "85d06533-3c0b-46a8-8fdb-379db9ae7e2e",
		Name:         "Assassin Gauntlet",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Assassin Gauntlet — draw a card, then discard a card",
					func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card, then discard a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(grantKey),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Assassin Gauntlet — attach it to up to one target creature you control; tap all creatures target opponent controls", assassinGauntletEnters),
				effects.Clauses(
					effects.UpToOneTargetCreature("up to one target creature you control", effects.YouControl()),
					effects.TargetPlayer("target opponent", effects.Opponent()),
				),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

func assassinGauntletEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
		if err := g.AttachSourceForEffect(item, mine); err != nil {
			return err
		}
	}
	opp, ok := ctx.ClauseTarget(1)
	if !ok || opp.Kind != game.TargetPlayer {
		return nil
	}
	return effects.B29TapAllCreaturesControlledBy(ctx, opp.ID)
}
