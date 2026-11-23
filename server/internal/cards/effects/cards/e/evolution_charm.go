package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolution Charm — Instant {1}{G} (Planar Chaos, #5595):
//
//	"Choose one —
//	 • Search your library for a basic land card, reveal it, put it into your
//	 hand, then shuffle.
//	 • Return target creature card from your graveyard to your hand.
//	 • Target creature gains flying until end of turn."
//
// Bushwhack's basic-land search, Darigaaz's Charm's graveyard return, and
// a flying grant until end of turn, each on its own bullet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f73fd4e-0d24-4daa-8d82-73c0cf7f801d",
		Name:         "Evolution Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Search your library for a basic land card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Evolution Charm — a basic land card",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTheModesGraveyardTargetToHand),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains flying until end of turn.",
				effects.TargetCreature("target creature"),
				effects.PlcGrantTheModesTarget("Evolution Charm — flying until end of turn", "flying")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))),
		),
	})
}
