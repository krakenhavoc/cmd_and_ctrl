package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devoted Duelist — Creature — Goblin Monk {1}{R}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Haste
//	 Flurry — Whenever you cast your second spell each turn, this
//	 creature deals 1 damage to each opponent."
//
// Flurry is an ability word (CR 207.2c): the trigger is Poised
// Practitioner's YouCastYourSecondSpellEachTurn, and the body one damage
// event from the Duelist to each opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f7620c46-b3d8-4694-ac7c-a43933c86c22",
		Name:            "Devoted Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Devoted Duelist — flurry: 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
