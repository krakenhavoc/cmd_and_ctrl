package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howl of the Horde — Sorcery {2}{R} (Khans of Tarkir):
//
//	"When you next cast an instant or sorcery spell this turn, copy that
//	 spell. You may choose new targets for the copy.
//	 Raid — If you attacked this turn, when you next cast an instant or
//	 sorcery spell this turn, copy that spell an additional time. You may
//	 choose new targets for the copy."
//
// Each sentence is its own delayed triggered ability (CR 603.7), Dual
// Strike's copy without the mana value ceiling; the raid one exists only if
// you attacked as this resolves (CR 608.2c), so both fire on the same cast
// and make two copies. Neither copies Howl itself: that cast was over before
// it resolved, and each ends with the turn (CR 603.7b).
//
// No simplification.
func init() {
	filter := game.CastFilter{Types: []string{"Instant", "Sorcery"}}
	effects.Register(effects.Spec{
		OracleID:     "76d9c310-9d06-43e8-87a9-fb03318a6cb8",
		Name:         "Howl of the Horde",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.WhenYouNextCast("Howl of the Horde — copy that spell; you may choose new targets for the copy",
				filter, effects.CopyTheSpellBody).Apply(ctx); err != nil {
				return err
			}
			if !effects.SpellRaid()(ctx.Game, item) {
				return nil
			}
			return effects.WhenYouNextCast("Howl of the Horde — raid: copy that spell an additional time; you may choose new targets for the copy",
				filter, effects.CopyTheSpellBody).Apply(ctx)
		},
	})
}
