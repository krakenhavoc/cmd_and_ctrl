package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clackbridge Troll — Creature — Troll {3}{B}{B}, 8/8 (Throne of Eldraine,
// S82 old sets):
//
//	"Trample, haste
//	 When this creature enters, target opponent creates three 0/1 white Goat
//	 creature tokens.
//	 At the beginning of combat on your turn, any opponent may sacrifice a
//	 creature of their choice. If a player does, tap this creature, you gain
//	 3 life, and you draw a card."
//
// The enters trigger is Balefang's: the targeted opponent creates the tokens
// (they are theirs). The combat trigger is Desecration Demon's any-opponent
// offer (ADR 0150), but only on its controller's turn: each opponent in APNAP
// order is asked until one sacrifices a creature; that pays for tapping the
// Troll, the Troll's controller gaining 3 life and drawing a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c2f25af-0b16-488d-9249-026a84bdb27f",
		Name:            "Clackbridge Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Clackbridge Troll — target opponent creates three 0/1 white Goat creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.CreateToken{Controller: victim, Template: effects.TokenCard("0/1 white Goat"), N: 3}.Apply(ctx)
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.AtBeginningOfYourCombat("Clackbridge Troll — any opponent may sacrifice a creature; if a player does, tap it, you gain 3 life and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AnyPlayerPays{
						Pay:       effects.SacrificePayment(1, "creature", "creatures", effects.QueryType("creature")),
						Opponents: true,
						Question:  "Clackbridge Troll — sacrifice a creature to tap it, and its controller gains 3 life and draws a card?",
						Then:      clackbridgeTrollTapped,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// clackbridgeTrollTapped is "if a player does", keyed. An on-disk key:
// never renamed, never reused.
var clackbridgeTrollTapped = effects.AnyPlayerConsequences("any-player/clackbridge-troll-tap",
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		if err := (effects.TapTarget{Target: ctx.Item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.GainLife{Player: ctx.Item.Controller, Amount: 3}).Apply(ctx); err != nil {
			return err
		}
		return effects.DrawCards{Player: ctx.Item.Controller, N: 1}.Apply(ctx)
	},
	nil,
	nil)
