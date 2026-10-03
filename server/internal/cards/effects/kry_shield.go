package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2c4bd475-b8af-4916-b7a0-68abb8994138",
		Name:         "Kry Shield",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Prevent all damage that would be dealt this turn by target creature you control. That creature gets +0/+X until end of turn, where X is its mana value.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return shieldAgainstTheTargetThen(NewContext(g, item), false, toughnessByManaValueUntilEOT)
			},
		}},
	})
}
