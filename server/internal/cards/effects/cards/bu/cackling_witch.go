package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cackling Witch — Creature — Human Spellshaper {1}{B}, 1/1 (Mercadian
// Masques, S82 OLD b682, #5619):
//
//	"{X}{B}, {T}, Discard a card: Target creature gets +X/+0 until end of
//	 turn."
//
// X is announced with the cost (CR 107.3k) and read back off the stack item;
// the pump is layer 7c until end of turn on any controller's creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee608650-410a-4d6b-be6e-63d7955cff39",
		Name:         "Cackling Witch",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{B}, {T}, Discard a card: Target creature gets +X/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.PowerXUntilEOT(0))},
			Cost:    effects.Plus(effects.ManaCost("{X}{B}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Cackling Witch — +X/+0"}.Apply(ctx)
			},
		}},
	})
}
