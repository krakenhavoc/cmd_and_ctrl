package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonstone Eulogist — Creature — Bat Warlock {3}{B}{B}, 4/4 (Commander,
// #5368):
//
//	"Flying
//	 Whenever a creature an opponent controls dies, you create a Blood
//	 token.
//	 Whenever you sacrifice an artifact, put a +1/+1 counter on this
//	 creature and you gain 1 life."
//
// The first trigger is Mari, the Killing Quill's death condition (the
// creature left the battlefield for a graveyard under an opponent's
// control, last-known, CR 603.10a) with the printed Blood token. The
// second fires once per artifact sacrificed (a Blood token's own cost
// counts: Crime Novelist's condition), putting the counter on the
// Eulogist and gaining the life in that order; a Eulogist that has left
// by resolution gets no counter, and the life is still gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf9190e0-6f53-44fb-8afa-c16f5bf9829e",
		Name:            "Moonstone Eulogist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B18OpponentsCreatureDied(ev, source, g)
			}, "Moonstone Eulogist — create a Blood token",
				effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})),
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B12YouSacrificedAnArtifact(ev, source, g)
			}, "Moonstone Eulogist — put a +1/+1 counter on it and gain 1 life",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
