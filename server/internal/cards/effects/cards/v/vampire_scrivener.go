package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Scrivener — Creature — Vampire Warlock {4}{B}, 2/2 (Streets of New
// Capenna, #5491):
//
//	"Flying
//	 Whenever you gain life during your turn, put a +1/+1 counter on this
//	 creature.
//	 Whenever you lose life during your turn, put a +1/+1 counter on this
//	 creature."
//
// Two printed abilities, so two registered triggers. Each fires once per
// life-change event, not once per point (a 3-life gain is one counter). The
// gain is a positive EventChangeLife (lifelink included, YouGainedLife); the
// loss is a negative one or damage that cost life (s22PlayerLostLife, which
// reads infect and "can't lose life" correctly). Both only count while its
// controller is the active player (IsYourTurn).
//
// No simplification.
func init() {
	gain := effects.On(game.EventChangeLife,
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return effects.IsYourTurn(g, source.Controller) && effects.YouGainedLife(ev, source, lki, g)
		},
		"Vampire Scrivener — you gained life during your turn: put a +1/+1 counter on it",
		effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
	lose := effects.OnAny([]game.EventKind{game.EventChangeLife, game.EventDealDamage},
		func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if !effects.IsYourTurn(g, source.Controller) {
				return false
			}
			_, lost := effects.S22PlayerLostLife(ev, source.Controller, g)
			return lost
		},
		"Vampire Scrivener — you lost life during your turn: put a +1/+1 counter on it",
		effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
	effects.Register(effects.Spec{
		OracleID:        "a945710e-fab6-4b54-ba40-586f2e54baf4",
		Name:            "Vampire Scrivener",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{gain, lose},
	})
}
