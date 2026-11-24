package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paul Atreides, Ducal Heir — Legendary Creature — Human Noble {1}{W}, 1/2
// (Secret Lair Drop):
//
//	"Vigilance
//	 Training
//	 {T}: Put a +1/+1 counter on each other creature you control. Activate
//	 only if Paul Atreides's power is 4 or greater."
//
// The power gate is read through the layered view when the ability is
// activated (CR 602.5b: a restriction on activation), not when it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c23ecf92-eda4-43b3-a5e7-7235b936b02b",
		Name:            "Paul Atreides, Ducal Heir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Training("Paul Atreides, Ducal Heir — training: put a +1/+1 counter on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on each other creature you control. Activate only if Paul Atreides's power is 4 or greater.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				c, ok := g.LookupCardForEffect(source)
				return ok && c.CurrentPower() >= 4
			},
			Effect: effects.PutCounterOnEachOtherCreatureYouControl,
		}},
	})
}
