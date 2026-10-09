package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lyra, Tolarian Archangel — Legendary Creature — Angel Wizard {1}{U}{U},
// 3/3:
//
//	"Flying
//	 At the beginning of each end step, if you've drawn three or more
//	 cards this turn, create a 3/3 blue Angel creature token with flying.
//	 {3}{U}{U}: Until end of turn, whenever Lyra deals combat damage to a
//	 player, draw two cards."
//
// The end-step trigger has an intervening "if" (CR 603.4): it is checked
// when the step begins and again as the ability resolves, against the
// per-turn draw tally. It fires at EVERY end step, so on an opponent's
// turn it counts what Lyra's controller has drawn that turn.
//
// The activated ability schedules a repeating delayed trigger (CR 603.7b,
// Benefactor's Draught's shape) keyed on Lyra herself: each combat-damage
// event to a player dealt by her until cleanup draws two. Activating it
// twice stacks two such triggers, as printed.
//
// No simplification.
func init() {
	const endLabel = "Lyra, Tolarian Archangel — create a 3/3 blue Angel with flying"
	Register(Spec{
		OracleID:        "d9060fff-0b65-4b3d-930d-6e36315b802c",
		Name:            "Lyra, Tolarian Archangel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return rfCreatureCDrewThreeOrMore(g, source.Controller)
			}, endLabel, func(g *game.Game, item *game.StackItem) error {
				if !rfCreatureCDrewThreeOrMore(g, item.Controller) {
					return nil
				}
				return CreateToken{Template: TokenCard("3/3 blue Angel with flying"), N: 1}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{U}{U}: Until end of turn, whenever Lyra deals combat damage to a player, draw two cards.",
			Cost:  ManaCost("{3}{U}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DelayedOnEvent{
					Label:     "Lyra, Tolarian Archangel — draw two cards",
					On:        []game.EventKind{game.EventDealDamage},
					Condition: rfCreatureCLyraCombatDamageCondition,
					Body:      rfCreatureCLyraDrawTwoBody,
					Repeats:   true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
