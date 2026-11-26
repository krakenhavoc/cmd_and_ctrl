package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lobber Crew — Creature — Goblin Warrior {2}{R}, 0/4 (Return to Ravnica,
// S82 OLD):
//
//	"Defender
//	 {T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast a multicolored spell, untap this creature."
//
// The tap ability pings every opponent. The untap trigger watches your own
// casts whose colours, read as the spell sits on the stack, number two or
// more (CR 105.2b), and untaps the Crew only while it is still the permanent
// that trigger came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84dffb4b-02d7-48e4-9a95-43f88cd52bf3",
		Name:            "Lobber Crew",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && effects.B25CastIsMulticolored(ev, g)
			}, "Lobber Crew — untap this creature", effects.UntapThisPermanentEffect),
		},
	})
}
