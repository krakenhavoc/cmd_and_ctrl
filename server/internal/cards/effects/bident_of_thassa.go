package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bident of Thassa — Legendary Enchantment Artifact for {2}{U}{U}:
//
//	"Whenever a creature you control deals combat damage to a player,
//	you may draw a card.
//	{1}{U}, {T}: Creatures your opponents control attack this turn if
//	able."
//
// S19 sub-PR 7: the first combat-damage trigger. EventDealDamage
// carries Combat=true for damage dealt in the combat damage step;
// combatDamageToPlayerBy checks the source is a creature the
// Bident's controller controls and the target is a player. One
// event per attacker per player, so three unblocked attackers queue
// three "you may draw" prompts — as in paper. The activated
// forced-attack ability waits for the activated-ability pipeline.
func init() {
	Register(Spec{
		OracleID:     "e1afaef7-9fa3-4662-a95f-adfb0da9fd11",
		Name:         "Bident of Thassa",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{1}{U}, {T}: Creatures your opponents control attack this turn if able\" ability isn't implemented — only the combat-damage draw trigger works."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bident of Thassa — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Bident of Thassa — draw a card?",
			},
		}},
	})
}
