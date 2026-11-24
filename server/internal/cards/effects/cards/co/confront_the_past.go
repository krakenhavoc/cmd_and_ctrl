package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confront the Past — Sorcery — Lesson {X}{B} (Strixhaven: School of Mages,
// #5502):
//
//	"Choose one —
//	 • Return target planeswalker card with mana value X or less from your
//	   graveyard to the battlefield.
//	 • Remove twice X loyalty counters from target planeswalker an opponent
//	   controls."
//
// Profane Command's X-bound graveyard target (WithManaValueAtMostX, judged on the
// card's mana value as it lies in the graveyard, and bound to the announced X
// both when the target is chosen and as the spell resolves). The second bullet
// removes 2X loyalty counters through the counter window (CR 122.8), a
// planeswalker with fewer loses what it has and then dies to the state-based
// action (CR 704.5i). X = 0 changes nothing, so the card declares XMatters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a92c1ffe-88b0-467c-883c-6ca3564a39e3",
		Name:         "Confront the Past",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target planeswalker card with mana value X or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target planeswalker card with mana value X or less from your graveyard", effects.YouOwn(), effects.Planeswalker()).WithManaValueAtMostX(),
				effects.ModeReturnTargetToTheBattlefield),
			effects.ModeDoing("Remove twice X loyalty counters from target planeswalker an opponent controls.",
				effects.TargetPermanent("target planeswalker an opponent controls", effects.Planeswalker(), effects.OpponentControls()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard || ctx.X() <= 0 {
						return nil
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterLoyalty, N: -2 * ctx.X()}.Apply(ctx)
				}),
		),
	})
}
