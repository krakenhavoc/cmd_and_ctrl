package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5f73fd4e-0d24-4daa-8d82-73c0cf7f801d",
		Name:         "Evolution Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Search your library for a basic land card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Evolution Charm — a basic land card",
					}.Apply(ctx)
				}),
			ModeDoing("Return target creature card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				ReturnTheModesGraveyardTargetToHand),
			ModeWithPurpose(ModeDoing("Target creature gains flying until end of turn.",
				TargetCreature("target creature"),
				plcGrantTheModesTarget("Evolution Charm — flying until end of turn", "flying")),
				ForTargets(BenefitUntilEOT(0, 0, 0, "flying"))),
		),
	})
}
