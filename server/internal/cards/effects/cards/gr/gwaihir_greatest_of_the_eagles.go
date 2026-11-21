package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gwaihir, Greatest of the Eagles — Legendary Creature — Bird Noble
// {4}{W}, 5/5 (Tales of Middle-earth Commander):
//
//	"Flying
//	 Whenever Gwaihir attacks, target attacking creature gains flying until
//	 end of turn.
//	 At the beginning of each end step, if you gained 3 or more life this
//	 turn, create a 3/3 white Bird creature token with flying and
//	 "Whenever this token attacks, target attacking creature gains flying
//	 until end of turn.""
//
// The end-step clause is on every player's end step and has an intervening
// "if" (CR 603.4), checked as the step begins and again as the ability
// resolves, over the life gained this turn (the net total gained, not net
// of life lost). The Bird is a catalog token template (ADR 0083), so its
// attack trigger travels with the token; the attack trigger shared with
// Gwaihir targets any attacking creature, Gwaihir or the Bird included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77572e70-4b8b-46ae-8c8d-f71b7dd456e0",
		Name:            "Gwaihir, Greatest of the Eagles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.S82GapB118GwaihirAttack("Gwaihir, Greatest of the Eagles"),
			effects.InterveningIf(effects.CondLifeGained{AtLeast: 3},
				effects.AtEachStep(game.StepEnd, "Gwaihir, Greatest of the Eagles — create a 3/3 white Bird creature token with flying and an attack trigger",
					effects.Do(effects.CreateToken{Template: effects.S82GapB118GwaihirBirdToken(), N: 1}))),
		},
	})
}
