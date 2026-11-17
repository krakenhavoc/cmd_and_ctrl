package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "164651c5-b831-4070-8491-16d30ac8395b",
		Name:         "Unite the Coalition",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose five (you may choose the same mode more than once)", 5, 5,
			ModeDoing("Target permanent phases out.",
				TargetPermanent("target permanent"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return PhaseOut{Targets: []uuid.UUID{t.ID}}.Apply(ctx)
				}),
			ModeDoing("Target player draws a card.", TargetPlayer("target player"), modeTargetPlayerDraws),
			ModeDoing("Exile target player's graveyard.", TargetPlayer("target player"), exileTheModesTargetPlayersGraveyard),
			ModeDoing("Unite the Coalition deals 2 damage to any target.", TargetAny(), DealFixedDamageToModesTarget(2)),
			ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				DestroyTheModesTarget),
		),
	})
}
