package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unite the Coalition — Instant {2}{W}{U}{B}{R}{G} (Dominaria United, S82 DMU
// batch 15):
//
//	"Choose five. You may choose the same mode more than once.
//	 • Target permanent phases out.
//	 • Target player draws a card.
//	 • Exile target player's graveyard.
//	 • Unite the Coalition deals 2 damage to any target.
//	 • Destroy target artifact or enchantment."
//
// Fiery Confluence's shape (CR 700.2d, ChooseNRepeating): exactly five picks
// from the five-bullet menu, any bullet repeatable, each occurrence with its
// own target (CR 601.2c) and running in announce order. Each occurrence is
// judged on its own as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "164651c5-b831-4070-8491-16d30ac8395b",
		Name:         "Unite the Coalition",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose five (you may choose the same mode more than once)", 5, 5,
			effects.ModeDoing("Target permanent phases out.",
				effects.TargetPermanent("target permanent"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.PhaseOut{Targets: []uuid.UUID{t.ID}}.Apply(ctx)
				}),
			effects.ModeDoing("Target player draws a card.", effects.TargetPlayer("target player"), effects.ModeTargetPlayerDraws),
			effects.ModeDoing("Exile target player's graveyard.", effects.TargetPlayer("target player"), effects.ExileTheModesTargetPlayersGraveyard),
			effects.ModeDoing("Unite the Coalition deals 2 damage to any target.", effects.TargetAny(), effects.DealFixedDamageToModesTarget(2)),
			effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.DestroyTheModesTarget),
		),
	})
}
