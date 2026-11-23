package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyship Buccaneer — Creature — Human Pirate {3}{U}{U}, 4/3
// (Foundations, #4465):
//
//	"Flying
//	 Raid — When this creature enters, if you attacked this turn, draw
//	 a card."
//
// Storm Fleet Spy's raid draw on a flier. Raid is an intervening "if"
// (CR 603.4): checked as it triggers and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5951f734-05d3-41ad-94ee-19bd310077d2",
		Name:            "Skyship Buccaneer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
				"Skyship Buccaneer — Raid: draw a card",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{effects.DrawCards{N: 1}}})),
		},
	})
}
