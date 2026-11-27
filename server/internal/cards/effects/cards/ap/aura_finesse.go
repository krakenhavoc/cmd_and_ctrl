package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aura Finesse — Instant {U} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Attach target Aura you control to target creature.
//	 Draw a card."
//
// Magnetic Theft's two-slot clause with an Aura you control for the first
// slot. The attach only writes the link (control never changes) and an Aura
// that could not legally enchant the creature stays where it is (CR 701.3b),
// the engine's own check. The draw is a separate step (CR 608.2c): a spell
// with one of its two targets gone still draws; with both gone it does not
// resolve at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "445c46d1-b571-4fe9-92cc-2873242d748f",
		Name:         "Aura Finesse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: effects.Clauses(
			effects.TargetPermanent("target Aura you control", effects.HasSubtype("Aura"), effects.YouControl()),
			effects.TargetCreature("target creature"),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.AttachClauseTargets(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
