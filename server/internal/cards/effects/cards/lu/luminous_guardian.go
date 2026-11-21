package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luminous Guardian — Creature — Human Nomad {3}{W}, 1/4:
//
//	"{W}: This creature gets +0/+1 until end of turn.
//	 {2}: This creature can block an additional creature this turn."
//
// Two activations on itself: a layer 7c pump (BoostUntilEOT) and
// Coastline Chimera's extra block (#1715).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d9a31f5-20d0-4266-95ec-38b4fa96b8aa",
		Name:         "Luminous Guardian",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}: This creature gets +0/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
				Cost:    effects.ManaCost("{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Target:    item.SourceCardID,
						Toughness: 1,
						Label:     "Luminous Guardian — +0/+1",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}: This creature can block an additional creature this turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{2}"),
				Effect:  effects.SelfBlocksAdditionalThisTurn("Luminous Guardian — can block an additional creature this turn"),
			},
		},
	})
}
