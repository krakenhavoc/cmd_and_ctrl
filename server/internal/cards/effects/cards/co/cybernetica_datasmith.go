package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cybernetica Datasmith — Artifact Creature — Human Artificer {1}{U}{B}, 0/1:
//
//	"Protection from Robots
//	 Field Reprogramming — {U}, {T}: Target player draws a card. Another
//	 target player creates a 4/4 colorless Robot artifact creature token with
//	 "This token can't block.""
//
// Two clauses: the second must be a different player (Distinct). Each part is
// judged on its own at resolution, so a player who has left costs only their
// own half (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23a265ef-d651-4d8d-94dd-0072f83541f9",
		Name:            "Cybernetica Datasmith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Robots"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target player draws a card. Another target player creates a 4/4 colorless Robot artifact creature token with \"This token can't block.\"",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Purpose: game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1},
			Targets: effects.Clauses(
				effects.TargetPlayer("target player"),
				effects.Distinct(effects.TargetPlayer("another target player")),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if t, ok := ctx.ClauseTarget(0); ok {
					if err := (effects.DrawCards{Player: t.ID, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				if t, ok := ctx.ClauseTarget(1); ok {
					return effects.CreateToken{Controller: t.ID, Template: effects.GapB190RobotToken(), N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
