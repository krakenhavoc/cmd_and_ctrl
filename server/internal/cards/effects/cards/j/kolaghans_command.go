package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kolaghan's Command — Instant {1}{B}{R} (EDHREC rank 1288):
//
//	"Choose two —
//	 • Return target creature card from your graveyard to your hand.
//	 • Target player discards a card.
//	 • Destroy target artifact.
//	 • Kolaghan's Command deals 2 damage to any target."
//
// The card ADR 0019 §8 named as the reason per-mode target slots had
// to exist, and the first "choose two" whose bullets each target.
// Until #764 it could not be declared at all: effects.Register
// panicked at boot on a Max > 1 spec with more than one targeted
// option, and the cast path refused an announcement that derived two
// clauses.
//
// Now each chosen bullet is its own mode OCCURRENCE with its own
// target group (TargetRef.Mode), the picker walks them in the order
// they were chosen, and resolution runs the chosen bullets in that
// same order (CR 608.2c). Choosing "destroy target artifact" and
// "deals 2 damage to any target" announces two targets, and each is
// re-checked against ITS OWN clause at resolution (CR 608.2b) — an
// artifact that stopped being one is skipped while the damage still
// happens.
//
// The discard is the printed one: the targeted player chooses which
// card (CR 701.9a), through the ordinary discard prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45f1e957-09f0-4d46-8e32-238f26060a87",
		Name:         "Kolaghan's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTheModesGraveyardTargetToHand),
			effects.ModeDoing("Target player discards a card.",
				effects.TargetPlayer("target player"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Kolaghan's Command — discard a card",
					})
					return nil
				}),
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Kolaghan's Command deals 2 damage to any target.",
				effects.TargetAny(),
				effects.DealFixedDamageToModesTarget(2)), effects.ForTargets(effects.DamageToTarget(0, 2))),
		),
	})
}
