package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Combine Guildmage — Creature — Merfolk Wizard {G}{U}, 2/2 (Ravnica
// Allegiance, #5521):
//
//	"{1}{G}, {T}: This turn, each creature you control enters with an
//	 additional +1/+1 counter on it.
//	 {1}{U}, {T}: Move a +1/+1 counter from target creature you control onto
//	 another target creature you control."
//
// The first ability is a replacement effect (CR 614.1c) held as a scoped
// record until cleanup, read live as each creature enters; a second
// resolution the same turn adds one more. The second moves the counter as
// Simic Guildmage does: it lands on the second creature first, and comes
// off the first only if it did (CR 122.5); the second target must be a
// different creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90abb55d-5c05-453f-b4a6-c0741dc36694",
		Name:         "Combine Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{G}, {T}: This turn, each creature you control enters with an additional +1/+1 counter on it.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.CreaturesEnterWithAnAdditionalPlusOneCounterThisTurnForEffect(item.SourceCardID, item.Controller,
						"Combine Guildmage — each creature you control enters with an additional +1/+1 counter this turn")
					return nil
				},
			},
			{
				Label:   "{1}{U}, {T}: Move a +1/+1 counter from target creature you control onto another target creature you control.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
				Targets: effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.Distinct(effects.TargetCreature("another target creature you control", effects.YouControl())),
				),
				Effect: effects.MoveAPlusOneCounterFromFirstToSecond,
			},
		},
	})
}
