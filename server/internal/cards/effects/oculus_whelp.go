package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "89ff1b0a-0cf3-455e-a381-1d93ef86a41b",
		Name:            "Oculus Whelp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key: key,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Oculus Whelp — draw a card", func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
			},
			Text: "When this creature dies, draw a card.",
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(
			func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && transformedPermanentsControlledBy(g, source.Controller) > 0
			}, key),
			"As long as you control a transformed permanent, this creature has \"When this creature dies, draw a card.\"")},
	})
}
