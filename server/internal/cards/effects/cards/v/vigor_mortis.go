package cardsv

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigor Mortis — Sorcery {2}{B}{B} (Duel Decks: Izzet vs. Golgari):
//
//	"Return target creature card from your graveyard to the battlefield. If
//	 {G} was spent to cast this spell, that creature enters with an additional
//	 +1/+1 counter on it."
//
// Recommission's reanimation with the counter as part of the entry (CR 614.1c),
// so counter doublers see it and its own enter triggers find it there. Whether
// green was spent is read off this spell's own payment (Once and Future's
// adamant read): any {G} counts, a cast with none does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6de27cba-5500-4448-abef-68e69fe519dc",
		Name:           "Vigor Mortis",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G"},
		Targets:        effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
		Purpose:        effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				var counters map[string]int
				if effects.AdamantSpent(ctx, "G", 1) {
					counters = map[string]int{game.CounterPlusOne: 1}
				}
				_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(t.ID, ctx.Controller(), false, counters)
				if err != nil && !errors.Is(err, game.ErrCardNotFound) {
					return err
				}
				return nil
			}
			return nil
		},
	})
}
