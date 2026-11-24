package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Said // Done — split card (Modern Horizons 2, #5501):
//
//	Said — Sorcery {2}{U}: "Return target instant or sorcery card from your
//	       graveyard to your hand."
//	Done — Instant {3}{U}: "Tap up to two target creatures. Those creatures
//	       don't untap during their controller's next untap step."
//
// Each half is an ADR 0034 face (ADR 0103): Said under the bare oracle ID,
// Done under "#1". Only one half is cast (CR 709.3; no fuse). Said is
// Corpse Hauler's return over an instant or sorcery card. Done is Will of the
// Naga's tap-and-freeze over up to two creatures (CR 502.3 for the missed
// untap step), each target judged as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d1d76e0-049c-4a33-96ec-8e60deed16d5",
		Name:         "Said",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "5d1d76e0-049c-4a33-96ec-8e60deed16d5#1",
		Name:         "Done",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Done")
		},
	})
}
