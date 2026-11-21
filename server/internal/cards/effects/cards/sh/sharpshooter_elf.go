package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sharpshooter Elf — Creature — Elf Ranger {2}{G}, */2 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Reach
//	 Sharpshooter Elf's power is equal to the number of creatures you control.
//	 When this creature enters, it deals damage equal to its power to target
//	 creature with flying an opponent controls."
//
// The power is Adeline's Layer 7a characteristic-defining ability (CR 604.3,
// CR 613.4a). The entry trigger is Murderous Redcap's: the damage is its power
// as the trigger resolves, last-known if it has left (CR 608.2h), named by
// object so deathtouch or lifelink it had are its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a25308ae-d757-4868-883d-62bbb00d6cbe",
		Name:            "Sharpshooter Elf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of creatures you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.B04CreaturesControlled(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Sharpshooter Elf — damage equal to its power to target creature with flying an opponent controls", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					elf, ok := ctx.TriggeringPermanent()
					if !ok || len(item.Targets) == 0 {
						return nil
					}
					ref := ctx.Trigger().Object.Ref()
					return effects.DealDamage{
						SourceObject: &ref,
						Target:       item.Targets[0].ID,
						Amount:       elf.Power,
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature with flying an opponent controls",
					effects.And(effects.HasKeyword("flying"), effects.OpponentControls())),
			),
		},
	})
}
