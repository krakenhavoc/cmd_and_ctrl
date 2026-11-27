package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boreal Centaur — Snow Creature — Centaur Warrior {1}{G}, 2/2 (Coldsnap, S82 OLD b560):
//
//	"{S}: This creature gets +1/+1 until end of turn. Activate only once each
//	 turn."
//
// {S} is paid with snow mana only (CR 107.4h); the once-a-turn limit counts
// activations (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{S}: This creature gets +1/+1 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "2b548f9a-c160-427d-9bb4-1fe940f8ab28",
		Name:         "Boreal Centaur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:      effects.ManaCost("{S}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 1, "Boreal Centaur — +1/+1 until end of turn"),
		}},
	})
}
