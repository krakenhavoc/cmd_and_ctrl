package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadewing Laureate — Creature — Human Warlock {W}{W/B}{B}, 2/2
// (Strixhaven: School of Mages, #5502):
//
//	"Flying
//	 Whenever another creature you control with flying dies, put a +1/+1
//	 counter on target creature you control."
//
// Whether the creature had flying is read from the keywords it last had
// (CR 603.10a, Event.WasKeyword), not from the card in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53a17d4f-cdbc-4669-a30e-5d5a334cc3c0",
		Name:            "Shadewing Laureate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AnotherCreatureYouControlDied(ev, source, g) {
					return false
				}
				was, known := ev.WasKeyword("flying")
				return known && was
			}, "Shadewing Laureate — put a +1/+1 counter on target creature you control", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
