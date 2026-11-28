package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wind Shear — Instant {2}{G} (Visions, S82 older sets batch 789, #5631):
//
//	"Attacking creatures with flying get -2/-2 and lose flying until end of
//	 turn."
//
// The set is locked as the spell resolves (CR 611.2c): the attacking
// creatures that have flying then, anyone's. One effect holds both changes, the
// shrink in layer 7c and the loss of flying in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa683059-b9de-4063-ae71-597ed4e2df30",
		Name:         "Wind Shear",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.IsAttackingCreature(), effects.HasKeyword("flying")),
				"Wind Shear — -2/-2 and loses flying", game.ModifyPTMod(-2, -2), game.RemoveKeywordsMod("flying"))
		},
	})
}
