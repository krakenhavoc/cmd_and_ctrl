package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bushwhack — Sorcery {G} (Foundations, #4465):
//
//	"Choose one —
//	 • Search your library for a basic land card, reveal it, put it into
//	   your hand, then shuffle.
//	 • Target creature you control fights target creature you don't
//	   control. (Each deals damage equal to its power to the other.)"
//
// The first bullet is Horizon Seeker's search; failing to find is legal
// and the library is shuffled either way. The second is Decisive Denial's
// fight: two clauses, the fight happening only while both are still legal
// (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c61374e5-a7f6-455e-a40b-a481751b536b",
		Name:         "Bushwhack",
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
						Reason:    "Bushwhack — a basic land card",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Target creature you control fights target creature you don't control.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				),
				effects.FightTheModesTargets),
		),
	})
}
