package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cori Mountain Stalwart — Creature — Human Monk {1}{R}{W}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flurry — Whenever you cast your second spell each turn, this
//	 creature deals 2 damage to each opponent and you gain 2 life."
//
// Flurry is an ability word (CR 207.2c): the trigger is Poised
// Practitioner's YouCastYourSecondSpellEachTurn. The body is Aurelia's
// drain: one damage event from the Stalwart to each opponent, then the
// life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "322a2d97-cf71-4bfc-b52d-f642f2d0dcbd",
		Name:         "Cori Mountain Stalwart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Cori Mountain Stalwart — flurry: 2 damage to each opponent and you gain 2 life",
				Do(DamageEachOpponent{N: 2}, GainLife{Amount: 2})),
		},
	})
}
