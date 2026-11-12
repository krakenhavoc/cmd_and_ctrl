package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Singing Tree — Creature — Plant {3}{G}, 0/3 (Arabian Nights, #5641):
//
//	"{T}: Target attacking creature has base power 0 until end of turn."
//
// A layer 7b base power until end of turn; toughness is untouched, and
// pumps and counters still apply on top of the 0. The target must be
// attacking as the ability is activated and again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23f9bee4-ac7e-4828-82c0-576fee0d29b7",
		Name:         "Singing Tree",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target attacking creature has base power 0 until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Effect:  basePowerZeroUntilEOT("Singing Tree — base power 0 until end of turn"),
		}},
	})
}

// basePowerZeroUntilEOT is "target creature has base power 0 until end
// of turn" (Singing Tree, Island of Wak-Wak): the first legal target's
// base power is set to 0 in layer 7b until cleanup, and its toughness is
// left alone.
func basePowerZeroUntilEOT(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return untilEndOfTurn(ctx, target, nil, label, game.SetBasePowerMod(0))
	}
}
