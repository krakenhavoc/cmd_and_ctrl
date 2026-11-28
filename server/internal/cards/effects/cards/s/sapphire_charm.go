package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sapphire Charm — Instant {U} (Mirage, S82 OLD b800, #5633):
//
//	"Choose one —
//	 • Target player draws a card at the beginning of the next turn's upkeep.
//	 • Target creature gains flying until end of turn.
//	 • Target creature an opponent controls phases out. (While it's phased
//	   out, it's treated as though it doesn't exist. It phases in before its
//	   controller untaps during their next untap step.)"
//
// Three bullets with their own target clauses. The draw is a delayed trigger
// at the next upkeep that names the player as data (CR 603.7); the flying is
// Raking Claws' keyword grant ending at cleanup; the phase-out is Divine
// Smite's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ea2bf6c-a427-4d0c-b447-159959327697",
		Name:         "Sapphire Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Target player draws a card at the beginning of the next turn's upkeep.", effects.TargetPlayer("target player")),
				game.Purpose{Answers: game.AnswerValue, Draws: 1}),
			effects.ModeWithPurpose(effects.Mode("Target creature gains flying until end of turn.", effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))),
			effects.ModeWithPurpose(effects.Mode("Target creature an opponent controls phases out.", effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				game.Purpose{Answers: game.AnswerRestrict}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.OldB800SapphireCharmDraw(item, ctx)
			case ctx.HasMode(1):
				return effects.IkoTargetGainsUntilEOT("flying", "Sapphire Charm — flying until end of turn")(item, ctx)
			case ctx.HasMode(2):
				return effects.OldB800PhaseOutTheTarget(item, ctx)
			}
			return nil
		},
	})
}
