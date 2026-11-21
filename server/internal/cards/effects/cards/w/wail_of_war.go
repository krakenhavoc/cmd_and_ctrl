package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wail of War — Instant {2}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one —
//	 • Creatures target opponent controls get -1/-1 until end of turn.
//	 • Return up to two target creature cards from your graveyard to your
//	   hand."
//
// The first mode is Neutralize the Guards' sweep: the creatures are those the
// opponent controls as the spell resolves (CR 611.2c), and an opponent who
// has left the game makes the mode's only target illegal (CR 608.2b). The
// second is Archenemy's Charm's return (eoeABCharmReturn): up to two, so the mode may name none, and
// each card still a creature card in the graveyard comes back; one that left
// in response stays where it went.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc0abdf3-3a88-4cb2-ba79-35fcaebbf7d2",
		Name:         "Wail of War",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Creatures target opponent controls get -1/-1 until end of turn.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.ControlledByPlayer(t.ID)), Power: -1, Toughness: -1,
						Label: "Wail of War — -1/-1",
					}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
			effects.ModeWithPurpose(effects.ModeDoing("Return up to two target creature cards from your graveyard to your hand.",
				effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
				effects.EoeABCharmReturn),
				game.Purpose{Answers: game.AnswerValue}),
		),
	})
}
