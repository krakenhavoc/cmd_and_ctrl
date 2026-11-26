package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bonfire of the Damned — Sorcery {X}{X}{R} (Avacyn Restored, S82 OLD, #5570):
//
//	"Bonfire of the Damned deals X damage to target player or planeswalker and
//	 each creature that player or that planeswalker's controller controls.
//	 Miracle {X}{R} (You may cast this card for its miracle cost when you draw
//	 it if it's the first card you drew this turn.)"
//
// Lavalanche's body on {X}{X}{R} (two X slots, so X is paid twice), with
// Entreat the Angels' {X} miracle price: X is chosen when the miracle cast is
// announced. The player is the target, or the controller of the targeted
// planeswalker, read before any damage lands; the creatures are snapshotted
// first and every point is one simultaneous damage instance (CR 608.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "13822dcf-0639-4852-9a7f-e1d55719e82c",
		Name:             "Bonfire of the Damned",
		Completeness:     effects.CompletenessFull,
		XMatters:         true,
		Targets:          effects.TargetPlayerOrPlaneswalker(),
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{X}{R}")},
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageXToTarget(0)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB437XDamageToPlayerOrWalkerAndTheirCreatures(ctx)
		},
	})
}
