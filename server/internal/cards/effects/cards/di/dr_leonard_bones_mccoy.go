package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dr. Leonard "Bones" McCoy — Legendary Creature — Human Doctor
// {1}{G}{G}, 3/2 (Star Trek, #4973):
//
//	"Flash
//	 When Dr. McCoy enters, choose one —
//	 • Return target non-Doctor permanent card from your graveyard to your
//	   hand.
//	 • Choose up to one other target creature. When it dies this turn,
//	   return that card to the battlefield under its owner's control with
//	   a +1/+1 counter on it at the beginning of the next end step."
//
// The first bullet re-checks its graveyard target at resolution (CR
// 608.2b); "non-Doctor" is read off the card as it is in the graveyard.
// The second creates a delayed trigger about the creature as the object
// it is now (CR 603.7a), ended at cleanup whether or not it fired (CR
// 603.7b). When it fires it creates a second delayed trigger about the
// card as the new object it is in the graveyard (CR 400.7), so a card
// that left the graveyard before the next end step is not returned (CR
// 603.7c). Any creature may be chosen, an opponent's included: the card
// comes back under its owner's control. A token that dies never comes
// back (CR 111.7).
//
// No simplification.
func init() {
	etb := effects.WhenThisEnters("Dr. McCoy — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = effects.ChooseOne(
		effects.ModeDoing("Return target non-Doctor permanent card from your graveyard to your hand.",
			effects.TargetCardInGraveyard("target non-Doctor permanent card from your graveyard", effects.YouOwn(), effects.Permanent(), effects.Not(effects.Subtype("Doctor"))),
			effects.NeoB08ReturnTheModesTargetToHand),
		effects.ModeDoing("Choose up to one other target creature. When it dies this turn, return that card to the battlefield under its owner's control with a +1/+1 counter on it at the beginning of the next end step.",
			effects.Another(effects.TargetCreature("up to one other target creature")).WithCount(0, 1),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard || t.ID == uuid.Nil {
					return nil
				}
				return effects.WhenThatCreatureDiesThisTurn(ctx, t.ID,
					"Dr. McCoy — when it dies this turn, return it at the beginning of the next end step", effects.TrekB20McCoyDiesBody)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "0157faee-f307-43c0-b308-bbf36ce02e32",
		Name:            "Dr. Leonard \"Bones\" McCoy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
