package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oculus Whelp — Creature — Phyrexian Dragon {3}{U}, 3/2 (March of the
// Machine):
//
//	"Flying
//	 As long as you control a transformed permanent, this creature has
//	 'When this creature dies, draw a card.'"
//
// The granted ability is a layer-6 bundle on the Whelp itself
// (GrantAbilities, ADR 0093), applied only while its controller controls a
// transformed permanent (CR 701.27g, TransformedPermanent). The grant is
// read on every layer pass, so the Whelp that dies while the condition
// holds has the trigger on the way out (CR 603.10a, leaves-the-battlefield
// abilities look back), and one that dies after the permanent has gone does
// not.
//
// No simplification.
func init() {
	const key = "oculus-whelp/dies-draw"
	effects.Register(effects.Spec{
		OracleID:        "89ff1b0a-0cf3-455e-a381-1d93ef86a41b",
		Name:            "Oculus Whelp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key: key,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Oculus Whelp — draw a card", func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
			},
			Text: "When this creature dies, draw a card.",
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(
			func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.TransformedPermanentsControlledBy(g, source.Controller) > 0
			}, key),
			"As long as you control a transformed permanent, this creature has \"When this creature dies, draw a card.\"")},
	})
}
