package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unexpected Potential — Conspiracy (Conspiracy, #5557):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 You may spend mana as though it were mana of any color to cast spells
//	 with the chosen name."
//
// Hidden agenda is the engine's (ADR 0227). Face up, a spend grant (CR 609.4b,
// Oath of Nissa's shape) read from the command zone as well as the
// battlefield: it covers a cast whose spell has the chosen name (the face being
// cast, CR 201.2) and no activation or other payment. The payment still uses
// each symbol's own colour when the pool has it, and {C} is not widened.
// Face down it has no abilities (CR 315.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1acbbf61-edf7-4f67-968b-6832c2dd54d1",
		Name:          "Unexpected Potential",
		Completeness:  effects.CompletenessFull,
		HiddenAgenda:  effects.HiddenAgendaOneName,
		AnyColorSpend: effects.YouMaySpendManaAsAnyColorToCastSpellsWithTheChosenName(),
	})
}
