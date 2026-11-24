package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Absorb Identity — Instant {1}{U}:
//
//	"Return target creature to its owner's hand. You may have Shapeshifters you
//	 control become copies of that creature until end of turn."
//
// The bounce is the first sentence; the "you may" is asked once it has settled
// (Gogo, Mysterious Mime's MayYes) and a yes makes every Shapeshifter its
// controller controls then (CR 611.2c: the set is locked as the answer arrives)
// a copy of the card that returned, the copiable values it has now (CR 707.2,
// Lazav's reading of a card that has moved), until end of turn. A changeling
// is a Shapeshifter; the bounced creature itself, now in a hand, is not on the
// battlefield to copy. A target that left or became illegal fizzles the whole spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4489f362-9b7e-4dc0-a459-a990a487454e",
		Name:         "Absorb Identity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.MayChoice{
				Question: "Absorb Identity — have Shapeshifters you control become copies of that creature until end of turn?",
				Then:     absorbIdentityAnswer,
				Carry:    []uuid.UUID{id},
			}.Apply(ctx)
		},
	})
}

// absorbIdentityAnswer is the "yes": each Shapeshifter its controller controls
// becomes a copy of the returned card until end of turn. Keyed (#4291).
var absorbIdentityAnswer = effects.MayYes("confirm/absorb-identity-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	of := effects.CarryAt(r, 0)
	if of == uuid.Nil {
		return nil
	}
	shapeshifter := ctx.TextSubtype("Shapeshifter")
	var targets []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype(shapeshifter) && c.InstanceID != of {
			targets = append(targets, c.InstanceID)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	return effects.BecomeCopy{Targets: targets, Of: of, Label: "Absorb Identity — becomes a copy"}.Apply(ctx)
})
