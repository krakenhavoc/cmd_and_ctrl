package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f7620c46-b3d8-4694-ac7c-a43933c86c22",
		Name:            "Devoted Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Devoted Duelist — flurry: 1 damage to each opponent",
				Do(DamageEachOpponent{N: 1})),
		},
	})
}
