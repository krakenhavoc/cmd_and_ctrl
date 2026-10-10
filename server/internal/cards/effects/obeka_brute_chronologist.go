package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obeka, Brute Chronologist — {1}{U}{B}{R} Legendary Creature — Ogre
// Wizard, 3/4:
//
//	"{T}: The player whose turn it is may end the turn. (Exile all
//	 spells and abilities from the stack. The player whose turn it is
//	 discards down to their maximum hand size. Damage wears off, and
//	 "this turn" and "until end of turn" effects end.)"
//
// ActivePlayerMayEndTheTurn (end_the_turn.go, CR 724.1, #2165): as the
// ability resolves, the ACTIVE player is asked, not Obeka's
// controller. Activated on your own turn it is Sundial of the Infinite
// for free; on an opponent's turn it hands that opponent a choice they
// will rarely take, which is the printed card. "No" leaves the turn
// alone.
//
// The {T} is a creature's tap cost, so Obeka cannot use it the turn it
// arrives (CR 302.6) — the engine enforces that for every creature's
// {T} ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21b76a94-d9f3-4c34-ab24-7e89321b04f0",
		Name:         "Obeka, Brute Chronologist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: The player whose turn it is may end the turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ActivePlayerMayEndTheTurn{}.Apply(NewContext(g, item))
			},
		}},
	})
}
