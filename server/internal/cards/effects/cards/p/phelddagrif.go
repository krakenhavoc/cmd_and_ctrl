package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phelddagrif — Legendary Creature — Phelddagrif {1}{G}{W}{U} 4/4 (Alliances,
// S82 OLD batch 819):
//
//	"{G}: Phelddagrif gains trample until end of turn. Target opponent creates
//	 a 1/1 green Hippo creature token.
//	 {W}: Phelddagrif gains flying until end of turn. Target opponent gains 2
//	 life.
//	 {U}: Return Phelddagrif to its owner's hand. Target opponent may draw a
//	 card."
//
// The three gifts go to the targeted opponent (Questing Phelddagrif's
// shapes); the opponent decides alone whether to draw.
func init() {
	opponent := func() *game.TargetSpec { return effects.TargetPlayer("target opponent", effects.Opponent()) }
	effects.Register(effects.Spec{
		OracleID:     "d0e99535-0ea0-4aa5-99f5-6c5255b27c38",
		Name:         "Phelddagrif",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}: Phelddagrif gains trample until end of turn. Target opponent creates a 1/1 green Hippo creature token.",
				Cost:    effects.ManaCost("{G}"),
				Targets: opponent(),
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Phelddagrif — trample"}).Apply(ctx); err != nil {
						return err
					}
					if opp := effects.TargetedPlayerOf(ctx); opp != uuid.Nil {
						return effects.CreateToken{Controller: opp, Template: effects.TokenCard("1/1 green Hippo"), N: 1}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{W}: Phelddagrif gains flying until end of turn. Target opponent gains 2 life.",
				Cost:    effects.ManaCost("{W}"),
				Targets: opponent(),
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Phelddagrif — flying"}).Apply(ctx); err != nil {
						return err
					}
					if opp := effects.TargetedPlayerOf(ctx); opp != uuid.Nil {
						return effects.GainLife{Player: opp, Amount: 2}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{U}: Return Phelddagrif to its owner's hand. Target opponent may draw a card.",
				Cost:    effects.ManaCost("{U}"),
				Targets: opponent(),
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// The opponent's offer is only queued here; it is answered after the
					// return below has been made, so the order is as printed.
					if opp := effects.TargetedPlayerOf(ctx); opp != uuid.Nil {
						if err := (effects.MayChoice{
							Player:   opp,
							Question: "Phelddagrif — draw a card?",
							Draws:    1,
							Then:     phelddagrifDraw,
						}).Apply(ctx); err != nil {
							return err
						}
					}
					if effects.OnBattlefield(g, item.SourceCardID) && !effects.SourceIsNewObject(g, item) {
						return effects.BounceToHand{Target: item.SourceCardID}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}

var phelddagrifDraw = effects.MayYes("confirm/phelddagrif-draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return effects.DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})
