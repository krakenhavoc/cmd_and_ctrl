package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dose of Dawnglow — Instant {4}{B} (Lorwyn Eclipsed, #4458):
//
//	"Return target creature card from your graveyard to the battlefield.
//	 Then if it isn't your main phase, blight 2. (Put two -1/-1 counters
//	 on a creature you control.)"
//
// The return is Raise the Past's simultaneous entry of one card, with the
// rest of the spell as its tail so it waits for any replacement prompt on
// the way in. "Your main phase" is a main phase of your own turn
// (b14InYourMainPhase), read as the tail runs. The blight is mandatory and
// is the card's effect, not a cost (Blight, CR 701.68a): you choose among
// the creatures you control that can have counters put on them, the
// returned one included, and with none you don't blight. With the target
// illegal the spell does nothing at all (CR 608.2b), so no blight.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f852bc33-e2f8-43be-9e6e-7badd1f894ef",
		Name:         "Dose of Dawnglow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{Targets: ids, ThenTail: doseOfDawnglowBlight}.Apply(ctx)
		},
	})
}

// doseOfDawnglowBlight is "then if it isn't your main phase, blight 2". An
// on-disk key.
var doseOfDawnglowBlight = effects.TogetherTail("dose-of-dawnglow/blight", func(ctx *effects.Context, _ []uuid.UUID, _ int) error {
	if effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
		return nil
	}
	return effects.Blight{N: 2}.Apply(ctx)
})
