package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knightfisher — Creature — Bird Knight {3}{U}{U}, 4/5 (Bloomburrow,
// #4467):
//
//	"Flying
//	 Whenever another nontoken Bird you control enters, create a 1/1
//	 blue Fish creature token."
//
// Lathliss's "another nontoken <type> you control enters"
// (AnotherNontokenCreatureOfTypeEnteredUnderYourControl), with "Bird"
// read through the Knightfisher's text (CR 612.2), so a changeling
// counts and a Bird token does not. The Fish is Fountainport's token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb2d222c-df42-41a7-a517-0a41e47a4280",
		Name:            "Knightfisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherNontokenCreatureOfTypeEnteredUnderYourControl("Bird"),
				"Knightfisher — create a 1/1 blue Fish creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Fish"), N: 1})),
		},
	})
}
