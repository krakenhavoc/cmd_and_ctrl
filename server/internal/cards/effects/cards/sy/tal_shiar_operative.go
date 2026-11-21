package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tal Shiar Operative — Creature — Romulan Officer Spy {1}{B}, 2/2 (Star
// Trek, #4973):
//
//	"Whenever this creature attacks, choose one —
//	 • Target creature you control gains deathtouch until end of turn.
//	 • Surveil 1."
//
// One modal attack trigger (Titanium Man's shape). The mode is chosen as
// the trigger is put on the stack (CR 700.2b), and only the first bullet
// has a target: with no creature you control to point at it can't be
// chosen, but the Operative is itself one while it attacks. The grant is
// fixed to the creature targeted (CR 611.2c); a target gone by resolution
// gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.WheneverThisAttacks("Tal Shiar Operative — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Target creature you control gains deathtouch until end of turn.",
			effects.TargetCreature("target creature you control", effects.YouControl()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				target, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: target.ID, Keywords: []string{"deathtouch"},
					Label: "Tal Shiar Operative — deathtouch"}.Apply(ctx)
			}),
		effects.ModeDoing("Surveil 1.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "53641cda-7093-4fed-97f1-380f771e00b0",
		Name:         "Tal Shiar Operative",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
