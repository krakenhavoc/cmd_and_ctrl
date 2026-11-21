package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fateshaper Aspirant — Creature — Rhino Cleric {4}{W}, 3/4:
//
//	"When this creature enters, choose one —
//	 • Return target legendary card from your graveyard to your hand.
//	 • Put a +1/+1 counter on target creature. It gains vigilance and
//	   indestructible until end of turn."
//
// A modal enters trigger (CR 603.3c): the bullet and its target are
// chosen as the trigger goes on the stack, and each bullet reads only
// its own target group. The graveyard bullet takes a legendary card of
// any type from the controller's own graveyard; the second bullet may
// target any creature, an opponent's included, as printed. A target
// that is gone at resolution is skipped (CR 608.2b), so a creature
// that left in response gets neither the counter nor the keywords.
//
// No simplification.
func init() {
	aspirant := effects.WhenThisEnters("Fateshaper Aspirant — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	aspirant.Modes = effects.ChooseOne(
		effects.ModeDoing("Return target legendary card from your graveyard to your hand.",
			effects.TargetCardInGraveyard("target legendary card in your graveyard", effects.YouOwn(), effects.Legendary()),
			effects.ReturnTheModesGraveyardTargetToHand),
		// The keywords wait for a prompt the counter asks (a CR 616 order; CR 608.2c, #4155).
		effects.ModeSteps("Put a +1/+1 counter on target creature. It gains vigilance and indestructible until end of turn.",
			effects.TargetCreature("target creature"),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"vigilance", "indestructible"},
					Label:    "Fateshaper Aspirant — vigilance and indestructible",
				}.Apply(ctx)
			})),
	)
	effects.Register(effects.Spec{
		OracleID:     "bb3996b6-b68d-4ea1-b3dd-1600f8da2f1c",
		Name:         "Fateshaper Aspirant",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{aspirant},
	})
}
