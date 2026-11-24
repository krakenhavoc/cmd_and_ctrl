package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Questing Phelddagrif — Creature — Phelddagrif {1}{G}{W}{U}, 4/4 (Secret Lair
// Drop):
//
//	"{G}: This creature gets +1/+1 until end of turn. Target opponent creates
//	 a 1/1 green Hippo creature token.
//	 {W}: This creature gains protection from black and from red until end of
//	 turn. Target opponent gains 2 life.
//	 {U}: This creature gains flying until end of turn. Target opponent may
//	 draw a card."
//
// Each ability does its own thing to the creature, then to the target
// opponent: the Hippo enters under the opponent's control, the life is theirs,
// and the draw is their choice (a prompt to that opponent).
func init() {
	opponent := func() *game.TargetSpec { return effects.TargetPlayer("target opponent", effects.Opponent()) }
	targetOpponent := effects.TargetedPlayerOf
	effects.Register(effects.Spec{
		OracleID:     "97f3afdb-dcee-44da-b210-c8d20030baff",
		Name:         "Questing Phelddagrif",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}: This creature gets +1/+1 until end of turn. Target opponent creates a 1/1 green Hippo creature token.",
				Cost:    effects.ManaCost("{G}"),
				Targets: opponent(),
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.ThisGetsUntilEndOfTurn(1, 1, "Questing Phelddagrif — +1/+1")(g, item); err != nil {
						return err
					}
					ctx := effects.NewContext(g, item)
					if opp := targetOpponent(ctx); opp != uuid.Nil {
						return effects.CreateToken{Controller: opp, Template: effects.TokenCard("1/1 green Hippo"), N: 1}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{W}: This creature gains protection from black and from red until end of turn. Target opponent gains 2 life.",
				Cost:    effects.ManaCost("{W}"),
				Targets: opponent(),
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"protection from black", "protection from red"},
						Label: "Questing Phelddagrif — protection from black and from red"}).Apply(ctx); err != nil {
						return err
					}
					if opp := targetOpponent(ctx); opp != uuid.Nil {
						return effects.GainLife{Player: opp, Amount: 2}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{U}: This creature gains flying until end of turn. Target opponent may draw a card.",
				Cost:    effects.ManaCost("{U}"),
				Targets: opponent(),
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
						Label: "Questing Phelddagrif — flying"}).Apply(ctx); err != nil {
						return err
					}
					opp := targetOpponent(ctx)
					if opp == uuid.Nil {
						return nil
					}
					return effects.MayChoice{
						Player:   opp,
						Question: "Questing Phelddagrif — draw a card?",
						Draws:    1,
						Then:     questingPhelddagrifDraw,
					}.Apply(ctx)
				},
			},
		},
	})
}

// questingPhelddagrifDraw is the opponent's answer: yes draws them a card.
// On-disk key.
var questingPhelddagrifDraw = effects.MayYes("confirm/questing-phelddagrif-draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return effects.DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})
