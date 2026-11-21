package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knife — Artifact — Clue Equipment {R} (Murders at Karlov Manor, #4469):
//
//	"During your turn, equipped creature gets +1/+0 and has first strike.
//	 {2}, Sacrifice this Equipment: Draw a card.
//	 Equip {2}"
//
// Hook Swords' statics at +1/+0: both are gated by Bilbo's Ring's
// duringSourceControllersTurn, read on every recompute, and "your" is the
// Equipment's controller (CR 301.5d). The draw is Candlestick's, printed
// on the card, so sacrificing Knife is sacrificing a Clue for every card
// that watches that. Equip is CR 702.6a.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7627163a-1689-446f-b5aa-a6c56aefb532",
		Name:         "Knife",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:    "During your turn, equipped creature gets +1/+0.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && effects.DuringSourceControllersTurn(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			},
			effects.GrantToAttachedWhile(effects.DuringSourceControllersTurn, "first strike"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this Equipment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			effects.EquipAbility("{2}"),
		},
	})
}
