package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Notorious Assassin — Creature — Human Spellshaper Assassin {3}{B}, 2/2
// (Mercadian Masques, S82 OLD, #5619):
//
//	"{2}{B}, {T}, Discard a card: Destroy target nonblack creature. It can't be
//	 regenerated."
//
// The discard is a cost, paid on activation, and the {T} is the Assassin's own, so
// it is summoning-sick the turn it arrives (CR 302.6). The target must be
// nonblack as the ability is activated and again as it resolves (CR 608.2b);
// regeneration shields are ignored (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "455138dd-4f76-4b40-9715-f1b1d61f8f23",
		Name:         "Notorious Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}, Discard a card: Destroy target nonblack creature. It can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target nonblack creature", effects.NonBlack()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
			},
		}},
	})
}
