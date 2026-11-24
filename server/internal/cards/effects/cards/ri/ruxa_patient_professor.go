package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruxa, Patient Professor — Legendary Creature — Bear Druid {2}{G}{G}, 4/4
// (Commander 2021, #5504):
//
//	"Whenever Ruxa enters or attacks, return target creature card with no
//	 abilities from your graveyard to your hand.
//	 Creatures you control with no abilities get +1/+1.
//	 For each creature you control with no abilities, you may have that
//	 creature assign its combat damage as though it weren't blocked."
//
// The trigger is Sun Titan's one-trigger-two-events shape with a target in
// the controller's graveyard: a creature CARD with no printed abilities
// (game.PrintsNoAbilities, the card's text off the battlefield, CR 109.3).
// The two statics read the battlefield creature as it is now
// (Card.HasNoAbilities: no keyword, no granted ability, and no printed
// ability an ability-removing effect has not taken away): the +1/+1 is a
// layer 7c modify, and the damage rule sets the creature's own "may assign
// its combat damage as though it weren't blocked" flag in layer 6 (CR
// 510.1c). Ruxa's text is an ability, so she is not covered.
//
// No simplification.
func init() {
	noAbilities := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && target.HasNoAbilities()
	}
	effects.Register(effects.Spec{
		OracleID:     "d760f2f0-6653-47c4-bd7b-aab2f398ddf2",
		Name:         "Ruxa, Patient Professor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Ruxa, Patient Professor — return target creature card with no abilities from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target creature card with no abilities from your graveyard", effects.YouOwn(), effects.B13NoAbilities())),
		},
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control with no abilities get +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: noAbilities,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			{
				Label:     "For each creature you control with no abilities, you may have that creature assign its combat damage as though it weren't blocked.",
				Layer:     game.Layer6Ability,
				AppliesTo: noAbilities,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.AssignsAsThoughUnblocked = true
				},
			},
		},
	})
}
