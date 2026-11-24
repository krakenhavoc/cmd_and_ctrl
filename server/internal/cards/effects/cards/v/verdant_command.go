package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Verdant Command — Instant {1}{G} (Modern Horizons 2):
//
//	"Choose two —
//	 • Target player creates two tapped 1/1 green Squirrel creature tokens.
//	 • Counter target loyalty ability of a planeswalker.
//	 • Exile target card from a graveyard.
//	 • Target player gains 3 life."
//
// Each bullet reads its own target group (CR 608.2b). The tokens and the life
// go to the player the bullet named, not the caster. The second bullet's
// clause is an ability on the stack whose source ability carries a loyalty
// cost (CR 606.2); a triggered ability or a creature's ability is not a legal
// target. The third reaches any graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a5d426b-e5eb-44f0-b5fe-c258bec6d59e",
		Name:         "Verdant Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Target player creates two tapped 1/1 green Squirrel creature tokens.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CreateTokenAdvanced{Controller: t.ID, Spec: effects.Token(effects.TokenCard("1/1 green Squirrel")).EntersTapped(), N: 2}.Apply(ctx)
				}),
			effects.ModeDoing("Counter target loyalty ability of a planeswalker.",
				effects.AbilityOnStack("target loyalty ability of a planeswalker", effects.OldB033LoyaltyAbilityOnStack()),
				effects.CounterTheModesTarget),
			effects.ModeDoing("Exile target card from a graveyard.",
				effects.TargetCardInGraveyard("target card from a graveyard"),
				effects.ExileTheModesTarget),
			effects.ModeDoing("Target player gains 3 life.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GainLife{Player: t.ID, Amount: 3}.Apply(ctx)
				}),
		),
	})
}
