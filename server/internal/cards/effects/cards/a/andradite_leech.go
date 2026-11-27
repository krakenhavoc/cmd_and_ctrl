package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Andradite Leech — Creature — Leech {2}{B}, 2/2 (Invasion, S82 OLD #5616):
//
//	"Black spells you cast cost {B} more to cast.
//	 {B}: This creature gets +1/+1 until end of turn."
//
// The tax is a cost increase over the controller's own black spells, one
// coloured {B} per spell (CR 601.2f), the Leech's own included once it is on
// the battlefield (it is not there while it is being cast). The pump is
// Firebreathing's shape on the Leech itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09bfc3b1-cc40-4cb9-87e4-86b0018caa75",
		Name:         "Andradite Leech",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.OldB651CostsMoreMana("{B}", "Black spells you cast cost {B} more to cast.", effects.YourSpell(), effects.ColoredSpell("B")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Andradite Leech — +1/+1 until end of turn"),
		}},
	})
}
