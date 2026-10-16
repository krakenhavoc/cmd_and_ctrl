package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Vapors — Sorcery {1}{U}:
//
//	"Lands target player controls don't untap during their next untap
//	 step."
//
// #3995's player-wide skip: read at that player's next untap step, so a
// land they play after Mana Vapors resolved stays tapped too, and one
// they stop controlling before then untaps as usual (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e082c95-cdc4-4f29-9915-05f83298f14d",
		Name:         "Mana Vapors",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return targetPlayersPermanentsDontUntap(ctx, "Mana Vapors — lands don't untap during that player's next untap step", "land")
		},
	})
}

// targetPlayersPermanentsDontUntap is "<types> target player controls
// don't untap during their next untap step", reading the player from the
// spell's still-legal player target (CR 608.2b).
func targetPlayersPermanentsDontUntap(ctx *Context, label string, types ...string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (PermanentsDontUntapNextStep{Player: t.ID, Types: types, Label: label}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
