package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodchief's Thirst — {B} Sorcery:
//
//	"Kicker {2}{B} (You may pay an additional {2}{B} as you cast this
//	 spell.)
//	 Destroy target creature or planeswalker with mana value 2 or less.
//	 If this spell was kicked, instead destroy target creature or
//	 planeswalker."
//
// The kicked clause is a whole different clause (#1716, WhenPaid): the
// kicker is announced at CR 601.2b, so an unkicked cast may not name a
// three-drop and a kicked one may name anything the printed type line
// allows.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4236851b-5366-43a1-bde4-f525b4fbcbce",
		Name:         "Bloodchief's Thirst",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{2}{B}"),
			effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())))},
		Targets: effects.TargetPermanent("target creature or planeswalker with mana value 2 or less",
			effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(2)),
		OnResolve: effects.S82MidDestroyTheTarget,
	})
}
