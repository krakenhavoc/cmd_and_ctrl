package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Paladin — Creature — Human Knight {3}{B}, 4/2 (Dominaria, #5530):
//
//	"Whenever you cast a historic spell, this creature deals 2 damage to each opponent. (Artifacts, legendaries, and Sagas are historic.)"
//
// Jhoira's historic cast trigger (CR 700.6) with Guttersnipe's payload:
// 2 damage to each opponent, resolved before the historic spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1104a746-169c-450a-80ad-bea0493c0576",
		Name:         "Cabal Paladin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastAHistoricSpell, "Cabal Paladin — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
		},
	})
}
