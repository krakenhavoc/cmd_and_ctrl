package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molten Note — Sorcery {X}{R}{W} (Secrets of Strixhaven, #4456):
//
//	"Molten Note deals damage to target creature equal to the amount of
//	 mana spent to cast this spell. Untap all creatures you control.
//	 Flashback {6}{R}{W}"
//
// The damage is the total mana paid for the spell (ReadsOwnPayment): X, the
// {R}{W} and anything else, so a flashback cast from the graveyard deals 8,
// the flashback cost being the mana spent (the rulings). A spell cast with
// no mana spent (a free cast) deals none. Untapping is a second step that
// does not depend on the damage; with its only target gone the spell does
// nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8c179b8f-3a83-49d6-9e7f-9e17c293ccbb",
		Name:             "Molten Note",
		Completeness:     effects.CompletenessFull,
		ReadsManaSpent:   game.ReadsOwnPayment,
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{6}{R}{W}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: ctx.ManaSpent().Total()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
			}),
		),
	})
}
