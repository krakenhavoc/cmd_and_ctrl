package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfbitten Captive // Krallenhorde Killer — Creature — Human Werewolf {G}
// 1/1 // Creature — Werewolf 2/2:
//
//	Wolfbitten Captive: "{1}{G}: This creature gets +2/+2 until end of turn.
//	 Activate only once each turn.
//	 At the beginning of each upkeep, if no spells were cast last turn,
//	 transform this creature."
//	Krallenhorde Killer: "{3}{G}: This creature gets +4/+4 until end of turn.
//	 Activate only once each turn.
//	 At the beginning of each upkeep, if a player cast two or more spells last
//	 turn, transform this creature."
func init() {
	const oracle = "7f59252f-3a37-4a22-938c-001f3208bd94"
	const frontPump = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	const backPump = "{3}{G}: This creature gets +4/+4 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Wolfbitten Captive",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     frontPump,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{G}"),
			Condition: effects.OncePerTurnActivation(frontPump),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Wolfbitten Captive — +2/+2 until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{effects.TransformAtUpkeepIfNoSpellsLastTurn("Wolfbitten Captive")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Krallenhorde Killer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     backPump,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{3}{G}"),
			Condition: effects.OncePerTurnActivation(backPump),
			Effect:    effects.ThisGetsUntilEndOfTurn(4, 4, "Krallenhorde Killer — +4/+4 until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Krallenhorde Killer")},
	})
}
