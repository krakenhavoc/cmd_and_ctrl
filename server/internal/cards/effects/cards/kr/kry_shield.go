package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kry Shield — Artifact {2}:
//
//	"{2}, {T}: Prevent all damage that would be dealt this turn by target creature you control. That creature gets +0/+X until end of turn, where X is its mana value."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7). All its
// damage, combat or not, is prevented for the rest of the turn; X is its
// mana value as the ability resolves.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c4bd475-b8af-4916-b7a0-68abb8994138",
		Name:         "Kry Shield",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Prevent all damage that would be dealt this turn by target creature you control. That creature gets +0/+X until end of turn, where X is its mana value.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent | game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ShieldAgainstTheTargetThen(effects.NewContext(g, item), false, effects.ToughnessByManaValueUntilEOT)
			},
		}},
	})
}
