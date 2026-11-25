package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beacon Bolt — Sorcery {1}{U}{R} (Guilds of Ravnica, #5525):
//
//	"Beacon Bolt deals damage to target creature equal to the total number of
//	 instant and sorcery cards you own in exile and in your graveyard.
//	 Jump-start"
//
// The count is read as the spell resolves, from the caster's graveyard and from
// the face-up exiled cards they own (Beacon Bolt itself is on the stack, so it
// is not counted; cast with jump-start it is exiled only after it resolves).
// Jump-start is the shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "884686dd-515c-4484-a324-e7d126903a42",
		Name:             "Beacon Bolt",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{1}{U}{R}")},
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			n := effects.InstantsAndSorceriesOwnedInExileAndGraveyard(ctx.Game, ctx.Controller())
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
		})),
	})
}
