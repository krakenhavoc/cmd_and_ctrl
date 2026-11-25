package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gandalf's Sanction — Sorcery {1}{U}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Gandalf's Sanction deals X damage to target creature, where X is the number
//	 of instant and sorcery cards in your graveyard. Excess damage is dealt to
//	 that creature's controller instead."
//
// X is counted as the spell resolves. Excess damage (CR 120.10) is what is
// beyond the lethal damage the creature still needs (toughness less the damage
// already marked), so the creature is dealt the lesser of X and that need and its
// controller is dealt the rest. The two are one damage event from the spell as its
// source. A creature whose toughness is not known has no lethal threshold, so all
// of it is dealt to the creature (the weaker reading).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6696d2e4-28ee-4ba7-8710-c1b735745270",
		Name:         "Gandalf's Sanction",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := 0
			if p := ctx.Game.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
				for _, gc := range p.Graveyard.Cards {
					if gc.IsInstant() || gc.IsSorcery() {
						x++
					}
				}
			}
			if x <= 0 {
				return nil
			}
			return effects.S82OldB116DamageToTargetCreatureExcessToItsController(item, ctx, x)
		},
	})
}
